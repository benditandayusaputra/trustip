package period

import (
	"strconv"
	"time"
)

func monthNames() [12]string {
	return [12]string{
		"Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
}

func DisplayDate(at time.Time) string {
	local := at.In(Jakarta())
	return strconv.Itoa(local.Day()) + " " + monthNames()[local.Month()-1] + " " + strconv.Itoa(local.Year())
}

const businessDateLayout = "2006-01-02"

func BusinessDate(at time.Time) string {
	return at.In(Jakarta()).Format(businessDateLayout)
}
