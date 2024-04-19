package moov

const (
	pathPing = "/ping"

	pathOAuth2Token  = "/oauth2/token"
	pathOAuth2Revoke = "/oauth2/revoke"

	pathAccounts = "/accounts"
	pathAccount  = "/accounts/%s"

	pathCapabilities = "/accounts/%s/capabilities"
	pathCapability   = "/accounts/%s/capabilities/%s"

	pathFiles = "/accounts/%s/files"
	pathFile  = "/accounts/%s/files/%s"

	pathPaymentMethods = "/accounts/%s/payment-methods"

	pathCards = "/accounts/%s/cards"
	pathCard  = "/accounts/%s/cards/%s"

	pathBankAccounts = "/accounts/%s/bank-accounts"
	pathBankAccount  = "/accounts/%s/bank-accounts/%s"

	pathBankAccountMicroDeposits = "/accounts/%s/bank-accounts/%s/microdeposits"

	pathWallets = "/accounts/%s/wallets"
	pathWallet  = "/accounts/%s/wallets/%s"

	pathWalletTransactions = "/accounts/%s/wallets/%s/transactions"
	pathWalletTransaction  = "/accounts/%s/wallets/%s/transactions/%s"

	pathApplePay        = "/accounts/%s/apple-pay"
	pathApplePayDomains = "/accounts/%s/apple-pay/domains"

	pathApplePaySessions = "/accounts/%s/apple-pay/sessions"
	pathApplePayTokens   = "/accounts/%s/apple-pay/tokens" // #nosec G101

	pathInstitutions = "/institutions/%s/search"

	pathTransferOptions = "/transfer-options"

	pathTransfers = "/transfers"
	pathTransfer  = "/transfers/%s"

	pathTransferReversals = "/transfers/%s/reversals"

	pathRefunds = "/transfers/%s/refunds"
	pathRefund  = "/transfers/%s/refunds/%s"

	pathDisputes = "/disputes"
	pathDispute  = "/disputes/%s"
)
