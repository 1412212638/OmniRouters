package billexport

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/shopspring/decimal"
)

const workbookNamespace = "http://schemas.openxmlformats.org/spreadsheetml/2006/main"
const workbookRelationshipsNamespace = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
const packageRelationshipsNamespace = "http://schemas.openxmlformats.org/package/2006/relationships"

type workbookSheet struct {
	name   string
	tables []workbookTable
}

type workbookTable struct {
	title string
	rows  [][]cell
	dates bool
}

var chineseWorkbookLabels = map[string]string{
	"Usage and refunds": "消耗与退款",
	"Wallet funding":    "钱包入账与调整",
	"Payments":          "订单支付",
	"By billing source": "按计费来源",
	"Other funding":     "其他入账",
	"Amounts follow the display unit on each row.": "请按每行单位核对金额和用量。",
	"No records in this view.":                     "此视图暂无记录。",
}

func workbookLabel(statement Statement, key string) string {
	if statement.Language == "zh" {
		if value, ok := chineseWorkbookLabels[key]; ok {
			return value
		}
	}
	return statement.label(key)
}

func workbookTableHeader(statement Statement, keys ...string) [][]cell {
	header := make([]cell, len(keys))
	for i, key := range keys {
		header[i] = str(workbookLabel(statement, key))
	}
	return [][]cell{header}
}

func workbookNonzero(value cell) cell {
	if value.numeric && value.value == "0" {
		return str("")
	}
	return value
}

type workbookUnitTotal struct {
	totals
	paid, gift, other    decimal.Decimal
	usage, funds, orders bool
}

func workbookOverview(statement Statement) workbookSheet {
	unitTotals := make(map[string]*workbookUnitTotal)
	for _, event := range statement.Events {
		total := unitTotals[event.Unit]
		if total == nil {
			total = &workbookUnitTotal{}
			unitTotals[event.Unit] = total
		}
		total.add(event)
		switch event.Kind {
		case Funding:
			total.funds = true
			switch event.Source {
			case "wallet_paid", "organization_topup":
				total.paid = total.paid.Add(event.Amount)
			case "wallet_gift":
				total.gift = total.gift.Add(event.Amount)
			default:
				total.other = total.other.Add(event.Amount)
			}
		case Adjustment:
			total.funds = true
		case Payment:
			total.orders = true
		default:
			total.usage = true
		}
	}

	units := make([]string, 0, len(unitTotals))
	for unit := range unitTotals {
		units = append(units, unit)
	}
	slices.Sort(units)
	if index := slices.Index(units, statement.quotaUnit()); index > 0 {
		unit := units[index]
		copy(units[1:index+1], units[:index])
		units[0] = unit
	}
	usage := workbookTable{
		title: "Usage and refunds",
		rows:  workbookTableHeader(statement, "Unit", "Net usage", "Charges", "Refunds", "Recorded savings", "Requests", "Unconfirmed usage (count)"),
	}
	funding := workbookTable{
		title: "Wallet funding",
		rows:  workbookTableHeader(statement, "Unit", "Funding", "Paid funding", "Gift funding", "Other funding", "Adjustments"),
	}
	payments := workbookTable{
		title: "Payments",
		rows:  workbookTableHeader(statement, "Unit", "Recorded payments", "Unverified payments (count)"),
	}
	for _, unit := range units {
		total := unitTotals[unit]
		label := str(statement.unitLabel(unit))
		if total.usage {
			savings := str("")
			if total.hasSavings {
				savings = amount(total.savings)
			}
			usage.rows = append(usage.rows, []cell{label, amount(total.net()), amount(total.charges), amount(total.refunds),
				savings, num(total.requests), num(total.unconfirmedUsage)})
		}
		if total.funds {
			funding.rows = append(funding.rows, []cell{label, amount(total.funding), amount(total.paid), amount(total.gift),
				amount(total.other), amount(total.adjustments)})
		}
		if total.orders {
			recorded := str("")
			if unit != UnverifiedCurrency {
				recorded = amount(total.payments)
			}
			payments.rows = append(payments.rows, []cell{label, recorded, num(total.unverifiedPayments)})
		}
	}
	var tables []workbookTable
	for _, table := range []workbookTable{usage, funding, payments} {
		if len(table.rows) > 1 {
			tables = append(tables, table)
		}
	}
	if len(tables) == 0 {
		tables = []workbookTable{usage}
	}
	return workbookSheet{name: statement.label("Overview"), tables: tables}
}

func workbookBreakdown(statement Statement, sheetKey, dimensionKey string, dimension func(Event) string) workbookSheet {
	byModel := dimensionKey == "Model"
	columns := []string{"Billing source", "Unit", "Net usage", "Charges", "Refunds", "Recorded savings"}
	if dimensionKey != "" {
		columns = append([]string{dimensionKey}, columns...)
	}
	if !byModel {
		columns = append(columns, "Funding", "Adjustments", "Recorded payments", "Unverified payments (count)")
	}
	columns = append(columns, "Requests", "Unconfirmed usage (count)", "Input tokens", "Output tokens", "Cache read tokens", "Cache write tokens")
	byDate := dimensionKey == "Posted date (UTC+08:00)"
	table := workbookTable{rows: workbookTableHeader(statement, columns...), dates: byDate}
	events := statement.Events
	if byModel {
		events = make([]Event, 0, len(statement.Events))
		for _, event := range statement.Events {
			if eventCategory(event.Kind) == usageCategory {
				events = append(events, event)
			}
		}
	}
	groups, keys := aggregate(events, func(event Event) (string, string) { return dimension(event), "" })
	for _, key := range keys {
		total := groups[key]
		var row []cell
		if dimensionKey != "" {
			dimensionCell := str(key.first)
			if byModel {
				dimensionCell = str(statement.billedModelName(key.first))
			}
			if byDate {
				if date, err := time.Parse("2006-01-02", key.first); err == nil {
					// Excel serial dates use an epoch 25,569 days before Unix time.
					dimensionCell = num(int(date.Unix()/86400) + 25569)
				}
			}
			row = append(row, dimensionCell)
		}
		savings := str("")
		if total.hasSavings {
			savings = amount(total.savings)
		}
		row = append(row, str(statement.codeLabel(key.source)), str(statement.unitLabel(key.unit)))
		usage := []cell{amount(total.net()), amount(total.charges), amount(total.refunds), savings}
		hasUsage := byModel || total.requests > 0 || total.unconfirmedUsage > 0 || !total.charges.IsZero() || !total.refunds.IsZero()
		if !hasUsage {
			usage = make([]cell, len(usage))
		}
		row = append(row, usage...)
		if !byModel {
			recorded := str("")
			if key.unit != UnverifiedCurrency {
				recorded = workbookNonzero(amount(total.payments))
			}
			row = append(row, workbookNonzero(amount(total.funding)), workbookNonzero(amount(total.adjustments)), recorded,
				workbookNonzero(num(total.unverifiedPayments)))
		}
		row = append(row, workbookNonzero(num(total.requests)), workbookNonzero(num(total.unconfirmedUsage)),
			workbookNonzero(amount(decimal.NewFromInt(total.inputTokens))), workbookNonzero(amount(decimal.NewFromInt(total.outputTokens))),
			workbookNonzero(amount(decimal.NewFromInt(total.cacheReadTokens))), workbookNonzero(amount(decimal.NewFromInt(total.cacheWriteTokens))))
		table.rows = append(table.rows, row)
	}
	return workbookSheet{name: workbookLabel(statement, sheetKey), tables: []workbookTable{table}}
}

const (
	workbookTitleStyle = iota + 1
	workbookRuleStyle
	workbookLabelStyle
	workbookNoteStyle
	workbookHeaderStyle
	workbookTextStyle
	workbookStripedTextStyle
	workbookIntegerStyle
	workbookStripedIntegerStyle
	workbookDecimalStyle
)

const workbookDecimalScales = 16
const workbookDateStyle = workbookDecimalStyle + 2*workbookDecimalScales

const workbookStyleTemplate = `<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
<numFmts count="%d">%s</numFmts>
<fonts count="5">
	<font><sz val="11"/><color rgb="FF243447"/><name val="Arial"/></font>
	<font><b/><sz val="16"/><color rgb="FF15334F"/><name val="Arial"/></font>
	<font><b/><sz val="11"/><color rgb="FF344D63"/><name val="Arial"/></font>
	<font><sz val="10"/><color rgb="FF52677A"/><name val="Arial"/></font>
	<font><b/><sz val="11"/><color rgb="FFFFFFFF"/><name val="Arial"/></font>
</fonts>
<fills count="4">
	<fill><patternFill patternType="none"/></fill>
	<fill><patternFill patternType="gray125"/></fill>
	<fill><patternFill patternType="solid"><fgColor rgb="FF234460"/><bgColor indexed="64"/></patternFill></fill>
	<fill><patternFill patternType="solid"><fgColor rgb="FFF1F6FA"/><bgColor indexed="64"/></patternFill></fill>
</fills>
<borders count="3">
	<border><left/><right/><top/><bottom/><diagonal/></border>
	<border><left/><right/><top/><bottom style="thin"><color rgb="FFA7BDCD"/></bottom><diagonal/></border>
	<border><left/><right style="thin"><color rgb="FF59758B"/></right><top/><bottom/><diagonal/></border>
</borders>
<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>
<cellXfs count="%d">
	<xf xfId="0" numFmtId="0" fontId="0" fillId="0" borderId="0" applyAlignment="1"><alignment vertical="center"/></xf>
	<xf xfId="0" numFmtId="0" fontId="1" fillId="0" borderId="0" applyFont="1" applyAlignment="1"><alignment vertical="center"/></xf>
	<xf xfId="0" numFmtId="0" fontId="0" fillId="0" borderId="1" applyBorder="1"/>
	<xf xfId="0" numFmtId="0" fontId="2" fillId="0" borderId="0" applyFont="1" applyAlignment="1"><alignment vertical="center"/></xf>
	<xf xfId="0" numFmtId="0" fontId="3" fillId="0" borderId="0" applyFont="1" applyAlignment="1"><alignment vertical="center"/></xf>
	<xf xfId="0" numFmtId="0" fontId="4" fillId="2" borderId="2" applyFont="1" applyFill="1" applyBorder="1" applyAlignment="1"><alignment horizontal="center" vertical="center" wrapText="1"/></xf>
	<xf xfId="0" numFmtId="0" fontId="0" fillId="0" borderId="0" applyAlignment="1"><alignment vertical="center"/></xf>
	<xf xfId="0" numFmtId="0" fontId="0" fillId="3" borderId="0" applyFill="1" applyAlignment="1"><alignment vertical="center"/></xf>
	<xf xfId="0" numFmtId="3" fontId="0" fillId="0" borderId="0" applyNumberFormat="1" applyAlignment="1"><alignment horizontal="right" vertical="center"/></xf>
	<xf xfId="0" numFmtId="3" fontId="0" fillId="3" borderId="0" applyNumberFormat="1" applyFill="1" applyAlignment="1"><alignment horizontal="right" vertical="center"/></xf>
%s
</cellXfs>
<cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles>
</styleSheet>`

func workbookStyles() string {
	var formats, styles strings.Builder
	for scale := 1; scale <= workbookDecimalScales; scale++ {
		formatID := 163 + scale
		fmt.Fprintf(&formats, `<numFmt numFmtId="%d" formatCode="#,##0.%s"/>`, formatID, strings.Repeat("0", scale))
		for _, fill := range []int{0, 3} {
			fmt.Fprintf(&styles, `<xf xfId="0" numFmtId="%d" fontId="0" fillId="%d" borderId="0" applyNumberFormat="1" applyFill="1" applyAlignment="1"><alignment horizontal="right" vertical="center"/></xf>`, formatID, fill)
		}
	}
	dateFormatID := 164 + workbookDecimalScales
	fmt.Fprintf(&formats, "<numFmt numFmtId=%q formatCode=%q/>", fmt.Sprint(dateFormatID), "yyyy-mm-dd")
	for _, fill := range []int{0, 3} {
		fmt.Fprintf(&styles, "<xf xfId=\"0\" numFmtId=\"%d\" fontId=\"0\" fillId=\"%d\" borderId=\"0\" applyNumberFormat=\"1\" applyFill=\"1\" applyAlignment=\"1\"><alignment horizontal=\"left\" vertical=\"center\"/></xf>", dateFormatID, fill)
	}
	return fmt.Sprintf(workbookStyleTemplate, workbookDecimalScales+1, formats.String(), workbookDateStyle+2, styles.String())
}

// Inline strings keep imported names as values and avoid formula evaluation.
func writeWorkbook(out io.Writer, statement Statement) error {
	sheets := []workbookSheet{
		workbookOverview(statement),
		workbookBreakdown(statement, "By billing source", "", func(Event) string { return "" }),
		workbookBreakdown(statement, "By model", "Model", func(event Event) string { return event.Model }),
		workbookBreakdown(statement, "By day", "Posted date (UTC+08:00)", func(event Event) string { return postedDate(event.PostedAt) }),
	}
	if statement.Organization {
		sheets = append(sheets, workbookBreakdown(statement, "By member", "Member ID", func(event Event) string { return optionalInt(event.UserID).value }))
	}

	archive := zip.NewWriter(out)
	var content, workbook, relationships strings.Builder
	fmt.Fprint(&content, `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>`)
	fmt.Fprintf(&workbook, `<workbook xmlns="%s" xmlns:r="%s"><bookViews><workbookView activeTab="0"/></bookViews><sheets>`, workbookNamespace, workbookRelationshipsNamespace)
	fmt.Fprintf(&relationships, `<Relationships xmlns="%s"><Relationship Id="rIdStyles" Type="%s/styles" Target="styles.xml"/>`, packageRelationshipsNamespace, workbookRelationshipsNamespace)
	for index, sheet := range sheets {
		id := index + 1
		path := fmt.Sprintf("worksheets/sheet%d.xml", id)
		fmt.Fprintf(&content, `<Override PartName="/xl/%s" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`, path)
		workbook.WriteString(`<sheet name="`)
		if err := xml.EscapeText(&workbook, []byte(sheet.name)); err != nil {
			return err
		}
		fmt.Fprintf(&workbook, `" sheetId="%d" r:id="rId%d"/>`, id, id)
		fmt.Fprintf(&relationships, `<Relationship Id="rId%d" Type="%s/worksheet" Target="%s"/>`, id, workbookRelationshipsNamespace, path)
		if err := writeWorkbookSheet(archive, "xl/"+path, statement, sheet, index == 0); err != nil {
			return err
		}
	}
	content.WriteString(`</Types>`)
	workbook.WriteString(`</sheets></workbook>`)
	relationships.WriteString(`</Relationships>`)
	for _, part := range []struct{ path, body string }{
		{"[Content_Types].xml", content.String()},
		{"_rels/.rels", fmt.Sprintf(`<Relationships xmlns="%s"><Relationship Id="rId1" Type="%s/officeDocument" Target="xl/workbook.xml"/></Relationships>`, packageRelationshipsNamespace, workbookRelationshipsNamespace)},
		{"xl/workbook.xml", workbook.String()},
		{"xl/_rels/workbook.xml.rels", relationships.String()},
		{"xl/styles.xml", workbookStyles()},
	} {
		if err := writeZIPText(archive, part.path, part.body); err != nil {
			return err
		}
	}
	return archive.Close()
}

func writeWorkbookSheet(archive *zip.Writer, path string, statement Statement, sheet workbookSheet, overview bool) error {
	const headerRow = 9
	widths := workbookWidths(statement, sheet)

	var body strings.Builder
	fmt.Fprintf(&body, `<worksheet xmlns="%s">`, workbookNamespace)
	if overview {
		body.WriteString(`<sheetPr><tabColor rgb="FF234460"/></sheetPr>`)
	}
	body.WriteString(`<sheetViews><sheetView showGridLines="0" workbookViewId="0">`)
	if overview {
		body.WriteString(`<selection activeCell="A1" sqref="A1"/>`)
	} else {
		fmt.Fprintf(&body, `<pane xSplit="1" ySplit="%d" topLeftCell="B%d" activePane="bottomRight" state="frozen"/><selection pane="topRight" activeCell="B1" sqref="B1"/><selection pane="bottomLeft" activeCell="A%d" sqref="A%d"/><selection pane="bottomRight" activeCell="B%d" sqref="B%d"/>`, headerRow, headerRow+1, headerRow+1, headerRow+1, headerRow+1, headerRow+1)
	}
	body.WriteString(`</sheetView></sheetViews>`)
	body.WriteString(`<sheetFormatPr defaultRowHeight="19"/><cols>`)
	for column, width := range widths {
		fmt.Fprintf(&body, `<col min="%d" max="%d" width="%d" customWidth="1"/>`, column+1, column+1, width)
	}
	body.WriteString(`</cols><sheetData>`)

	title := sheet.name
	if overview {
		title = "OmniRouters " + statement.label("Billing summary")
	}
	period := statement.StartMonth
	if statement.StartMonth != statement.EndMonth {
		period += " – " + statement.EndMonth
	}
	metadata := []struct {
		index  int
		height int
		style  int
		values []cell
	}{
		{2, 28, workbookTitleStyle, []cell{str(title)}},
		{3, 8, workbookRuleStyle, make([]cell, len(widths))},
		{4, 19, workbookLabelStyle, []cell{str(statement.label("Billing period")), str(period)}},
		{5, 19, workbookLabelStyle, []cell{str(statement.label("Account")), str(statement.Account)}},
		{6, 19, workbookLabelStyle, []cell{str(statement.label("Data through (UTC+08:00)")), str(postedTime(statement.AsOf.Unix()))}},
		{7, 19, workbookLabelStyle, []cell{str(statement.label("Display basis")), str(statement.displayBasis())}},
	}
	for _, row := range metadata {
		fmt.Fprintf(&body, `<row r="%d" ht="%d" customHeight="1">`, row.index, row.height)
		for column, value := range row.values {
			style := row.style
			if column > 0 && row.style == workbookLabelStyle {
				style = workbookTextStyle
			}
			if err := writeWorkbookCell(&body, column, row.index, value, style); err != nil {
				return err
			}
		}
		body.WriteString(`</row>`)
	}

	nextRow := headerRow
	for _, table := range sheet.tables {
		if table.title != "" {
			if err := writeWorkbookLabelRow(&body, nextRow, workbookLabel(statement, table.title), workbookLabelStyle); err != nil {
				return err
			}
			nextRow++
		}
		if err := writeWorkbookTable(&body, table, nextRow); err != nil {
			return err
		}
		nextRow += len(table.rows) + 1
	}
	if len(sheet.tables) == 1 && len(sheet.tables[0].rows) == 1 {
		if err := writeWorkbookLabelRow(&body, nextRow-1, workbookLabel(statement, "No records in this view."), workbookNoteStyle); err != nil {
			return err
		}
	}
	if overview && len(statement.Events) != 0 {
		for _, note := range []string{
			"Amounts follow the display unit on each row.",
			"Net usage is charges less refunds. Funding and adjustments are listed separately.",
			"Payments with unverified currency are listed by count.",
		} {
			if err := writeWorkbookLabelRow(&body, nextRow, workbookLabel(statement, note), workbookNoteStyle); err != nil {
				return err
			}
			nextRow++
		}
	}
	body.WriteString(`</sheetData>`)
	if !overview {
		rows := sheet.tables[0].rows
		fmt.Fprintf(&body, `<autoFilter ref="A%d:%s%d"/>`, headerRow, workbookColumnName(len(rows[0])-1), headerRow+len(rows)-1)
	}
	body.WriteString(`</worksheet>`)
	return writeZIPText(archive, path, body.String())
}

func workbookWidths(statement Statement, sheet workbookSheet) []int {
	var widths []int
	for _, table := range sheet.tables {
		headers := table.rows[0]
		if len(headers) > len(widths) {
			widths = append(widths, make([]int, len(headers)-len(widths))...)
		}
		for column, heading := range headers {
			width := max(13, min(22, workbookTextWidth(heading.value)+4))
			if table.dates && column == 0 {
				width = max(width, 20)
			}
			for _, row := range table.rows[1:] {
				value := row[column]
				if value.numeric {
					width = max(width, min(38, len(value.value)+len(value.value)/3+2))
				} else {
					width = max(width, min(36, workbookTextWidth(value.value)+2))
				}
			}
			widths[column] = max(widths[column], width)
		}
	}
	// The metadata labels are in column A on every sheet.
	widths[0] = max(widths[0], workbookTextWidth(statement.label("Data through (UTC+08:00)"))+2)
	return widths
}

func workbookTextWidth(value string) int {
	width := utf8.RuneCountInString(value)
	for _, char := range value {
		if char >= 0x2E80 {
			width++
		}
	}
	return width
}

func writeWorkbookLabelRow(out *strings.Builder, row int, label string, style int) error {
	fmt.Fprintf(out, `<row r="%d">`, row)
	if err := writeWorkbookCell(out, 0, row, str(label), style); err != nil {
		return err
	}
	out.WriteString(`</row>`)
	return nil
}

func writeWorkbookTable(out *strings.Builder, table workbookTable, firstRow int) error {
	for index, row := range table.rows {
		rowNumber := firstRow + index
		if index == 0 {
			fmt.Fprintf(out, `<row r="%d" ht="34" customHeight="1">`, rowNumber)
		} else {
			fmt.Fprintf(out, `<row r="%d">`, rowNumber)
		}
		for column, value := range row {
			style := workbookCellStyle(value, index)
			if table.dates && column == 0 && index != 0 && value.numeric {
				style = workbookDateStyle
				if index%2 == 0 {
					style++
				}
			}
			if err := writeWorkbookCell(out, column, rowNumber, value, style); err != nil {
				return err
			}
		}
		out.WriteString(`</row>`)
	}
	return nil
}

func workbookCellStyle(value cell, row int) int {
	if row == 0 {
		return workbookHeaderStyle
	}
	striped := row%2 == 0
	if !value.numeric {
		if striped {
			return workbookStripedTextStyle
		}
		return workbookTextStyle
	}
	if _, fractional, ok := strings.Cut(value.value, "."); ok && len(fractional) != 0 {
		// Go's decimals keep their exact scale. Formatting each value with that
		// scale avoids exposing artifacts from a workbook reader's binary float.
		style := workbookDecimalStyle + 2*(min(len(fractional), workbookDecimalScales)-1)
		if striped {
			return style + 1
		}
		return style
	}
	if striped {
		return workbookStripedIntegerStyle
	}
	return workbookIntegerStyle
}

func workbookColumnName(column int) string {
	var name []byte
	for index := column + 1; index > 0; index = (index - 1) / 26 {
		name = append([]byte{byte('A' + (index-1)%26)}, name...)
	}
	return string(name)
}

func writeWorkbookCell(out *strings.Builder, column, row int, value cell, style int) error {
	reference := fmt.Sprintf("%s%d", workbookColumnName(column), row)
	switch {
	case value.numeric:
		fmt.Fprintf(out, `<c r="%s" s="%d"><v>%s</v></c>`, reference, style, value.value)
	case value.value == "":
		fmt.Fprintf(out, `<c r="%s" s="%d"/>`, reference, style)
	default:
		fmt.Fprintf(out, `<c r="%s" s="%d" t="inlineStr"><is><t xml:space="preserve">`, reference, style)
		if err := xml.EscapeText(out, []byte(value.value)); err != nil {
			return err
		}
		out.WriteString(`</t></is></c>`)
	}
	return nil
}

func writeZIPText(archive *zip.Writer, path, body string) error {
	w, err := archive.Create(path)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, xml.Header+body)
	return err
}
