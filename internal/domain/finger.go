package domain

// Finger identifies the finger used to press a physical key.
// The assignment is standard touch typing, based only on the physical
// position of the key (row + column), independent of the character.
type Finger string

const (
	FingerLeftPinky   Finger = "LP"
	FingerLeftRing    Finger = "LR"
	FingerLeftMiddle  Finger = "LM"
	FingerLeftIndex   Finger = "LI"
	FingerRightIndex  Finger = "RI"
	FingerRightMiddle Finger = "RM"
	FingerRightRing   Finger = "RR"
	FingerRightPinky  Finger = "RP"
	FingerThumb       Finger = "TH"
)

// FingerFor returns the finger that presses the key at (row, index).
// Row is one of "numeric", "upper", "middle", "bottom"; index is 0-based
// left to right. It follows the standard ISO touch-typing columns.
func FingerFor(row string, index int) Finger {
	switch row {
	case "numeric":
		switch index {
		case 0, 1:
			return FingerLeftPinky
		case 2:
			return FingerLeftRing
		case 3:
			return FingerLeftMiddle
		case 4, 5:
			return FingerLeftIndex
		case 6, 7:
			return FingerRightIndex
		case 8:
			return FingerRightMiddle
		case 9:
			return FingerRightRing
		default:
			return FingerRightPinky
		}
	case "upper":
		switch index {
		case 0:
			return FingerLeftPinky
		case 1:
			return FingerLeftRing
		case 2:
			return FingerLeftMiddle
		case 3, 4:
			return FingerLeftIndex
		case 5, 6:
			return FingerRightIndex
		case 7:
			return FingerRightMiddle
		case 8:
			return FingerRightRing
		default:
			return FingerRightPinky
		}
	case "middle":
		switch index {
		case 0:
			return FingerLeftPinky
		case 1:
			return FingerLeftRing
		case 2:
			return FingerLeftMiddle
		case 3, 4:
			return FingerLeftIndex
		case 5, 6:
			return FingerRightIndex
		case 7:
			return FingerRightMiddle
		case 8:
			return FingerRightRing
		default:
			return FingerRightPinky
		}
	case "bottom":
		switch index {
		case 0:
			return FingerLeftPinky
		case 1:
			return FingerLeftRing
		case 2:
			return FingerLeftMiddle
		case 3, 4:
			return FingerLeftIndex
		case 5, 6:
			return FingerRightIndex
		case 7:
			return FingerRightMiddle
		case 8:
			return FingerRightRing
		default:
			return FingerRightPinky
		}
	default:
		return FingerLeftPinky
	}
}
