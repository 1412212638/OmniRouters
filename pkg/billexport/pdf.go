package billexport

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

//go:embed fonts/BillingSansSC.ttf
var billingSansSC []byte

var billingFont, billingFontError = sfnt.Parse(billingSansSC)

const (
	pdfInk    = "0.08 0.15 0.22"
	pdfMuted  = "0.30 0.38 0.45"
	pdfAccent = "0.08 0.40 0.37"
	pdfLine   = "0.84 0.89 0.91"
	pdfTint   = "0.93 0.97 0.97"
	pdfWhite  = "1 1 1"
)

type pdfGlyph struct {
	width int
	rune  rune
}

type statementPage struct {
	strings.Builder
	glyphs map[sfnt.GlyphIndex]pdfGlyph
}

func (page *statementPage) rect(x, y, width, height int, color string) {
	fmt.Fprintf(page, "%s rg %d %d %d %d re f\n", color, x, y, width, height)
}

func (page *statementPage) rule(y int) {
	fmt.Fprintf(page, "%s RG 0.6 w 42 %d m 553 %d l S\n", pdfLine, y, y)
}

var escapePDF = strings.NewReplacer("\\", "\\\\", "(", "\\(", ")", "\\)", "\n", " ", "\r", " ")

func (page *statementPage) text(x, y, size int, face, color, value string) {
	if strings.IndexFunc(value, func(r rune) bool { return r > 127 }) == -1 {
		fmt.Fprintf(page, "%s rg BT /%s %d Tf %d %d Td (%s) Tj ET\n", color, face, size, x, y, escapePDF.Replace(value))
		return
	}
	// Match the CJK font's metrics to the embedded glyph IDs and preserve Unicode
	// for copying/searching. Chinese PDF labels are drawn from the fixed catalog.
	fmt.Fprintf(page, "q %s rg %s RG 0.16 w BT /CJK %d Tf %d %d Td 2 Tr <", color, color, size, x, y)
	var buffer sfnt.Buffer
	for _, char := range value {
		glyph, _ := billingFont.GlyphIndex(&buffer, char)
		if glyph == 0 {
			char = '?'
			glyph, _ = billingFont.GlyphIndex(&buffer, char)
		}
		if _, seen := page.glyphs[glyph]; !seen {
			width, _ := billingFont.GlyphAdvance(&buffer, glyph, fixed.I(1000), font.HintingNone)
			page.glyphs[glyph] = pdfGlyph{width.Round(), char}
		}
		fmt.Fprintf(page, "%04X", glyph)
	}
	page.WriteString("> Tj ET Q\n")
}

// Keep dense monetary columns readable; the spreadsheet and CSV retain the
// precise recorded decimals.
func (page *statementPage) value(right, y int, n decimal.Decimal) {
	value := n.Round(6).String()
	size := min(10, 50*5/(3*len(value)))
	color := pdfInk
	if n.IsZero() {
		color = pdfMuted
	}
	page.text(right-(len(value)*size*3+4)/5, y, size, "N", color, value)
}

type statementSummary struct {
	requests, unconfirmed, records int
	walletUnit, tokens    decimal.Decimal
}

func (statement Statement) pdfCurrencySymbol() string {
	var buffer sfnt.Buffer
	for _, char := range statement.CurrencySymbol {
		glyph, _ := billingFont.GlyphIndex(&buffer, char)
		if glyph == 0 {
			return ""
		}
	}
	return statement.CurrencySymbol
}

func (statement Statement) pdfUnitLabel(unit string) string {
	if unit == UnitCustomCurrency {
		if symbol := statement.pdfCurrencySymbol(); symbol != "" {
			return statement.label("Custom") + " (" + symbol + ")"
		}
	}
	return statement.label(unit)
}

func layoutStatementPage(statement Statement, groups map[grouping]*totals, keys []grouping, summary statementSummary, pageNumber, pages int, glyphs map[sfnt.GlyphIndex]pdfGlyph) string {
	page := statementPage{glyphs: glyphs}
	label := statement.label
	page.rect(0, 833, 595, 9, pdfAccent)
	page.text(42, 792, 14, "B", pdfAccent, "OmniRouters")
	page.text(42, 751, 29, "B", pdfInk, label("Billing statement"))
	page.text(525, 792, 12, "B", pdfInk, "L0")
	page.rule(730)
	metadataLabels := []string{"Account", "Billing period", "Data through (UTC+08:00)"}
	metadataValues := []string{statement.Account, statement.StartMonth + " - " + statement.EndMonth, postedTime(statement.AsOf.Unix())}
	if statement.StartMonth == statement.EndMonth {
		metadataValues[1] = statement.StartMonth
	}
	positions := []int{42, 209, 380}
	for i, key := range metadataLabels {
		page.text(positions[i], 710, 9, "R", pdfMuted, label(key))
		page.text(positions[i], 690, 12, "B", pdfInk, metadataValues[i])
	}
	displayUnit := statement.pdfUnitLabel(statement.quotaUnit())
	displayBasis := label("Display basis") + ": " + displayUnit
	rateUnit := displayUnit
	if symbol := statement.pdfCurrencySymbol(); symbol != "" && statement.quotaUnit() == UnitCustomCurrency {
		rateUnit = symbol
	}
	if rate := statement.displayRate(rateUnit); rate != "" {
		displayBasis += " | " + rate
	}
	page.text(42, 676, 8, "R", pdfMuted, displayBasis)
	metricLabels := []string{label("Net usage") + " (" + displayUnit + ")", label("Net usage") + " (" + UnitTokens + ")", label("Recorded requests")}
	metricValues := []string{summary.walletUnit.Round(6).String(), summary.tokens.Round(6).String(), strconv.Itoa(summary.requests)}
	for i, title := range metricLabels {
		x := 42 + i*173
		page.rect(x, 603, 165, 64, pdfTint)
		page.text(x+13, 647, 9, "R", pdfMuted, title)
		value := metricValues[i]
		page.text(x+13, 619, min(23, 216/len(value)), "B", pdfInk, value)
	}

	page.text(42, 569, 14, "B", pdfInk, label("Reconciliation by billing source"))
	page.rect(42, 525, 511, 28, pdfInk)
	columns := []string{"Billing source", "Unit", "Funding", "Adjustments", "Charges", "Refunds", "Net usage"}
	left := []int{51, 185, 284, 337, 399, 451, 499}
	right := []int{325, 381, 437, 490, 546}
	for i, key := range columns {
		page.text(left[i], 536, 9, "B", pdfWhite, label(key))
	}
	if len(keys) == 0 {
		page.text(51, 498, 10, "R", pdfMuted, label("No billing records for this period."))
		page.rule(487)
	}
	for row, key := range keys {
		y := 525 - 25*row
		if row%2 == 0 {
			page.rect(42, y-25, 511, 25, pdfTint)
		}
		source, knownSource := billingCodes[key.source]
		if !knownSource {
			source = "Other"
		}
		page.text(51, y-16, 9, "R", pdfInk, label(source))
		page.text(185, y-16, 8, "R", pdfMuted, statement.pdfUnitLabel(key.unit))
		total := groups[key]
		if key.unit == UnverifiedCurrency {
			page.text(284, y-16, 9, "R", pdfMuted, fmt.Sprintf("%s: %d", label("Unverified payments (count)"), total.unverifiedPayments))
			continue
		}
		if !total.payments.IsZero() {
			page.text(284, y-16, 9, "R", pdfInk, label("Recorded payments")+": "+total.payments.Round(6).String())
			continue
		}
		values := []decimal.Decimal{total.funding, total.adjustments, total.charges, total.refunds, total.net()}
		for i, value := range values {
			page.value(right[i], y-16, value)
		}
	}
	if len(keys) > 0 {
		page.rule(525 - 25*len(keys))
	}

	page.rule(162)
	page.text(42, 146, 10, "B", pdfInk, label("Reconciliation notes"))
	for i, key := range []string{
		"Net usage is charges less refunds. Funding and adjustments are listed separately.",
		"Unconfirmed usage is excluded from charges.",
		"Payments with unverified currency are listed by count.",
		"This statement reflects retained records. Tax invoices are issued separately.",
	} {
		page.text(42, 129-14*i, 8, "R", pdfMuted, label(key))
	}
	page.rule(68)
	page.text(42, 49, 8, "R", pdfMuted, fmt.Sprintf("%s: %d   |   %s: %d", label("Billing records"), summary.records, label("Unconfirmed usage (count)"), summary.unconfirmed))
	page.text(481, 49, 8, "R", pdfMuted, fmt.Sprintf("%s %d / %d", label("Page"), pageNumber, pages))
	return page.String()
}

func pdfUnicodeMap(glyphs map[sfnt.GlyphIndex]pdfGlyph, ids []sfnt.GlyphIndex) string {
	var cmap strings.Builder
	cmap.WriteString("/CIDInit /ProcSet findresource begin\n12 dict begin\nbegincmap\n")
	cmap.WriteString("/CIDSystemInfo << /Registry (Adobe) /Ordering (UCS) /Supplement 0 >> def\n")
	cmap.WriteString("/CMapName /BillingSansSC-UCS def\n/CMapType 2 def\n1 begincodespacerange\n<0000> <FFFF>\nendcodespacerange\n")
	for start := 0; start < len(ids); start += 100 {
		batch := ids[start:min(start+100, len(ids))]
		fmt.Fprintf(&cmap, "%d beginbfchar\n", len(batch))
		for _, id := range batch {
			fmt.Fprintf(&cmap, "<%04X> <%04X>\n", id, glyphs[id].rune)
		}
		cmap.WriteString("endbfchar\n")
	}
	cmap.WriteString("endcmap\nCMapName currentdict /CMap defineresource pop\nend\nend\n")
	return cmap.String()
}

func writePDF(out io.Writer, statement Statement) error {
	if billingFontError != nil {
		return billingFontError
	}
	groups, keys := aggregate(statement.Events, nil)
	var summary statementSummary
	for key, total := range groups {
		summary.requests += total.requests
		summary.unconfirmed += total.unconfirmedUsage
		summary.records += total.records
		switch key.unit {
		case statement.quotaUnit():
			summary.walletUnit = summary.walletUnit.Add(total.net())
		case UnitTokens:
			summary.tokens = summary.tokens.Add(total.net())
		}
	}
	const rowsPerPage = 13
	pages := max(1, (len(keys)+rowsPerPage-1)/rowsPerPage)
	glyphs := make(map[sfnt.GlyphIndex]pdfGlyph)
	pageContents := make([]string, pages)
	for index := range pages {
		batch := keys[index*rowsPerPage : min((index+1)*rowsPerPage, len(keys))]
		pageContents[index] = layoutStatementPage(statement, groups, batch, summary, index+1, pages, glyphs)
	}

	var document bytes.Buffer
	document.WriteString("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")
	var offsets []int
	object := func(body string) {
		offsets = append(offsets, document.Len())
		fmt.Fprintf(&document, "%d 0 obj\n%s\nendobj\n", len(offsets), body)
	}
	stream := func(attributes, content string) {
		object(fmt.Sprintf("<< /Length %d %s >>\nstream\n%s\nendstream", len(content), attributes, content))
	}
	firstPage := 6
	if len(glyphs) > 0 {
		firstPage += 5
	}
	var children strings.Builder
	for index := range pages {
		fmt.Fprintf(&children, "%d 0 R ", firstPage+index*2)
	}
	object("<< /Type /Catalog /Pages 2 0 R >>")
	object(fmt.Sprintf("<< /Type /Pages /Count %d /Kids [%s] >>", pages, children.String()))
	object("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")
	object("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >>")
	object("<< /Type /Font /Subtype /Type1 /BaseFont /Courier >>")
	fontResources := "/R 3 0 R /B 4 0 R /N 5 0 R"
	if len(glyphs) > 0 {
		fontResources += " /CJK 6 0 R"
		ids := make([]sfnt.GlyphIndex, 0, len(glyphs))
		for id := range glyphs {
			ids = append(ids, id)
		}
		slices.Sort(ids)
		var widths strings.Builder
		for _, id := range ids {
			fmt.Fprintf(&widths, "%d [%d] ", id, glyphs[id].width)
		}
		object("<< /Type /Font /Subtype /Type0 /BaseFont /OMNIBI+BillingSansSC /Encoding /Identity-H /DescendantFonts [7 0 R] /ToUnicode 10 0 R >>")
		object(fmt.Sprintf("<< /Type /Font /Subtype /CIDFontType2 /BaseFont /OMNIBI+BillingSansSC /CIDSystemInfo << /Registry (Adobe) /Ordering (Identity) /Supplement 0 >> /FontDescriptor 8 0 R /CIDToGIDMap /Identity /W [%s] >>", widths.String()))
		object("<< /Type /FontDescriptor /FontName /OMNIBI+BillingSansSC /Flags 4 /FontBBox [-991 -1050 2930 1810] /ItalicAngle 0 /Ascent 880 /Descent -120 /CapHeight 733 /StemV 65 /FontFile2 9 0 R >>")
		stream(fmt.Sprintf("/Length1 %d", len(billingSansSC)), string(billingSansSC))
		stream("", pdfUnicodeMap(glyphs, ids))
	}
	for index, content := range pageContents {
		object(fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << %s >> >> /Contents %d 0 R >>", fontResources, firstPage+index*2+1))
		stream("", content)
	}
	xref := document.Len()
	fmt.Fprintf(&document, "xref\n0 %d\n0000000000 65535 f \n", len(offsets)+1)
	for _, offset := range offsets {
		fmt.Fprintf(&document, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&document, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets)+1, xref)
	_, err := document.WriteTo(out)
	return err
}
