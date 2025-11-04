package testengine

import (
	"github.com/hhruszka/secretscanner/checkers"
	"go.uber.org/zap"
)

const MIN_LENGHT = 4

// IsSecret checks if a word is a potential secret
func IsSecret(candidate string) bool {
	switch {
	case checkers.PasswordCheckerIsTimeOrDate(candidate):
		zap.L().Debug("PasswordCheckerIsTimeOrDate", zap.String("candidate", candidate))
		return false
	case checkers.PasswordCheckerIsFlag(candidate):
		zap.L().Debug("PasswordCheckerIsFlag", zap.String("candidate", candidate))
		return false
	case !checkers.PasswordCheckerIsASCII(candidate):
		zap.L().Debug("PasswordCheckerIsASCII", zap.String("candidate", candidate))
		return false
	case checkers.PasswordCheckerSymbolsAndDigitsOnly(candidate):
		zap.L().Debug("PasswordCheckerSymbolsAndDigitsOnly", zap.String("candidate", candidate))
		return false
	case checkers.PasswordCheckerDigitsOnly(candidate):
		zap.L().Debug("PasswordCheckerDigitsOnly", zap.String("candidate", candidate))
		return false
	case checkers.PasswordCheckerIsTCP_URI(candidate):
		zap.L().Debug("PasswordCheckerIsTCP_URI", zap.String("candidate", candidate))
		return false
	case checkers.PasswordCheckerIsHTTP_URI(candidate):
		zap.L().Debug("PasswordCheckerIsHTTP_URI", zap.String("candidate", candidate))
		return false
	case checkers.PasswordCheckerIsKubernetesInternalDomain(candidate):
		zap.L().Debug("PasswordCheckerIsKubernetesInternalDomain", zap.String("candidate", candidate))
		return false
	case checkers.PasswordCheckerIsEnglishSentence(candidate):
		zap.L().Debug("PasswordCheckerIsEnglishSentence", zap.String("candidate", candidate))
		return false
	case checkers.PasswordCheckerTooShort(candidate, MIN_LENGHT):
		zap.L().Debug("PasswordCheckerTooShort", zap.String("candidate", candidate))
		return false
	}
	return true
}
