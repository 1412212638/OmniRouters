package billexport

func (statement Statement) label(key string) string {
	if statement.Language == "zh" {
		if value, ok := chineseLabels[key]; ok {
			return value
		}
	}
	return key
}

func (statement Statement) header(keys ...string) []cell {
	row := make([]cell, len(keys))
	for index, key := range keys {
		row[index] = str(statement.label(key))
	}
	return row
}

func (statement Statement) codeLabel(code string) string {
	if label, ok := billingCodes[code]; ok {
		return statement.label(label)
	}
	return code
}

func (statement Statement) billedModelName(model string) string {
	if model == "" {
		return statement.label("Unrecorded model")
	}
	return model
}

func (statement Statement) quotaUnit() string {
	if statement.DisplayUnit == "" {
		return UnitUSD
	}
	return statement.DisplayUnit
}

func (statement Statement) unitLabel(unit string) string {
	label := statement.label(unit)
	if unit == UnitCustomCurrency && statement.CurrencySymbol != "" {
		return label + " (" + statement.CurrencySymbol + ")"
	}
	return label
}

func (statement Statement) displayRate(unit string) string {
	if statement.ExchangeRate == "" || (statement.DisplayUnit != UnitCNY && statement.DisplayUnit != UnitCustomCurrency) {
		return ""
	}
	return "1 USD = " + statement.ExchangeRate + " " + unit
}

func (statement Statement) displayBasis() string {
	unit := statement.quotaUnit()
	basis := statement.unitLabel(unit)
	rateUnit := unit
	if unit == UnitCustomCurrency && statement.CurrencySymbol != "" {
		rateUnit = statement.CurrencySymbol
	}
	if rate := statement.displayRate(rateUnit); rate != "" {
		return basis + " · " + rate
	}
	return basis
}

var billingCodes = map[string]string{
	Payment: "Payment", Funding: "Funding", Adjustment: "Adjustment",
	Charge: "Charge", Precharge: "Precharge", AdditionalCharge: "Additional charge",
	Refund: "Refund", UnconfirmedUsage: "Unconfirmed usage",
	"wallet": "Wallet", "subscription": "Subscription", "organization": "Organization wallet",
	"wallet_topup": "Top-up order", "subscription_order": "Subscription order",
	"wallet_paid": "Paid funding", "wallet_gift": "Gift funding",
	"wallet_adjustment": "Wallet adjustment", "wallet_unclassified": "Unclassified funding",
	"wallet_transfer_out": "Transfer to organization", "wallet_tool_fee": "Wallet tool fee",
	"organization_topup": "Organization top-up", "organization_transfer_in": "Organization transfer in",
	"unverified_funding_source": "Unverified billing source",
}

var chineseLabels = map[string]string{
	"Billing statement":                               "电子账单",
	"Billing summary":                                 "账单汇总",
	"Billing period":                                  "账单周期",
	"Account":                                         "账单账户",
	"Data through (UTC+08:00)":                        "数据截至（UTC+08:00）",
	"Billing records":                                 "账单记录",
	"Recorded requests":                               "记录请求数",
	"Unconfirmed usage":                               "待确认用量",
	"Reconciliation by billing source":                "按计费来源对账",
	"Reconciliation notes":                            "对账说明",
	"No billing records for this period.":             "本账期暂无账单记录。",
	"Amounts are grouped by billing source and unit.": "账单按计费来源和单位汇总。",
	"Payments with unverified currency are listed by count.":                            "币种未确认的支付单独列示笔数。",
	"Net usage is charges less refunds. Funding and adjustments are listed separately.": "净消耗为扣费减退款。入账与调整单独列示。",
	"This statement reflects retained records. Tax invoices are issued separately.":     "本账单基于已保留记录整理。发票另行开具。",
	"Unconfirmed usage is excluded from charges.":                                       "待确认用量未计入扣费。",
	"Page":                          "页",
	"Overview":                      "账单概览",
	"By model":                      "按模型",
	"By day":                        "按日期",
	"By member":                     "按成员",
	"Posted at (UTC+08:00)":         "入账时间（UTC+08:00）",
	"Posting month":                 "入账月份",
	"Posted date (UTC+08:00)":       "入账日期（UTC+08:00）",
	"Used at (UTC+08:00)":           "使用时间（UTC+08:00）",
	"First usage (UTC+08:00)":       "首次使用（UTC+08:00）",
	"Last usage (UTC+08:00)":        "末次使用（UTC+08:00）",
	"Latest posting (UTC+08:00)":    "最近入账（UTC+08:00）",
	"Event":                         "记录类型",
	"Billing source":                "计费来源",
	"Unit":                          "单位",
	"Posted amount":                 "入账数值",
	"Original amount":               "折前用量费用",
	"Savings":                       "折扣减免",
	"Recorded savings":              "已记录折扣减免",
	"Task precharge":                "任务预扣",
	"Task settled total":            "任务结算总额",
	"Model":                         "模型",
	"Unrecorded model":              "未记录模型",
	"API key ID":                    "密钥 ID",
	"API key name":                  "密钥名称",
	"Member ID":                     "成员 ID",
	"Request ID":                    "请求 ID",
	"Task ID":                       "任务 ID",
	"Order or entry reference":      "订单或流水编号",
	"Payment provider":              "支付渠道",
	"Recorded payments":             "已确认币种的支付",
	"Unverified payments (count)":   "币种未确认支付笔数",
	"Funding":                       "资金入账",
	"Adjustment":                    "资金调整",
	"Adjustments":                   "资金调整",
	"Charges":                       "扣费",
	"Refunds":                       "退款",
	"Net usage":                     "净消耗",
	"Display basis":                 "账单显示单位",
	"Custom":                        "自定义",
	UnitCustomCurrency:              "自定义货币",
	UnitQuota:                       "额度单位",
	"Requests":                      "请求数",
	"Unconfirmed usage (count)":     "待确认用量记录数",
	"Input tokens":                  "输入 Tokens",
	"Output tokens":                 "输出 Tokens",
	"Cache read tokens":             "缓存读取 Tokens",
	"Cache write tokens":            "缓存写入 Tokens",
	"Recorded task usage":           "任务用量",
	"Task usage unit":               "任务用量单位",
	UnverifiedCurrency:              "币种未确认",
	"Unverified subscription units": "订阅单位未确认",
	"Unverified usage units":        "用量单位未确认",
	"Seconds":                       "秒",
	"Payment":                       "支付",
	"Charge":                        "消费",
	"Precharge":                     "预扣",
	"Additional charge":             "补扣",
	"Refund":                        "退款",
	"Wallet":                        "个人钱包",
	"Subscription":                  "订阅套餐",
	"Organization wallet":           "组织钱包",
	"Top-up order":                  "充值订单",
	"Subscription order":            "订阅订单",
	"Paid funding":                  "付费入账",
	"Gift funding":                  "赠送入账",
	"Wallet adjustment":             "钱包调整",
	"Unclassified funding":          "未分类入账",
	"Transfer to organization":      "转账至组织",
	"Wallet tool fee":               "钱包工具费",
	"Organization top-up":           "组织充值入账",
	"Organization transfer in":      "组织转入",
	"Unverified billing source":     "计费来源未确认",
	"Other":                         "其他",
	"Personal account":              "个人账户",
	"Organization":                  "组织",
	"Record category":               "账目分类",
	"Usage":                         "用量",
	"Net usage change":              "净消耗变动",
	"Funding change":                "余额入账",
	"Adjustment amount":             "调整金额",
	"Recorded order amount":         "订单记录金额",
	"Unconfirmed amount":            "待确认金额",
	"Usage ended at (UTC+08:00)":    "使用结束（UTC+08:00）",
	"Last usage end (UTC+08:00)":    "末次使用结束（UTC+08:00）",
	"Ledger entry ID":               "流水 ID",
}
