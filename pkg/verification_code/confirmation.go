package verificationcode

func VerificationСodeСonfirmation(verificationСode string, userEnteredСode string) bool {
	if verificationСode == userEnteredСode {
		return true
	}
	return false
}
