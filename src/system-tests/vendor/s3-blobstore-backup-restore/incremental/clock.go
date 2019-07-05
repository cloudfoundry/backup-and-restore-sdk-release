package incremental

import "time"

type clock struct {
}

func (c clock) Now() string {
	return time.Now().Format("2006_01_02_15_04_05")
}
