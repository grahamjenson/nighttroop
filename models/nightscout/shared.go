package nightscout

import "fmt"

func ValidateOffset(offsetMinutes FlexibleInt64) error {
	if offsetMinutes < -1440 || offsetMinutes > 1440 {
		return fmt.Errorf("invalid offset: %d", offsetMinutes)
	}
	return nil
}

func ValidateDate(date DateInt) error {
	if date == 0 {
		return fmt.Errorf("invalid date: %d", date)
	}
	if date <= 100000000000 {
		return fmt.Errorf("invalid date: %d", date)
	}
	return nil
}
