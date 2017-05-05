package probe

import (
	"net/url"
	"strconv"
	"time"
	"fmt"
)

type ProbeRecord struct {
	Timestamp  int64        `json:"timestamp"`
	Mac        string       `json:"mac"`
	SequenceId int          `json:"sequence_id"`
	Rssi       int          `json:"rssi"`
	ApName     string       `json:"ap_name"`
}

func (record *ProbeRecord) Values() (values url.Values) {
	values.Add("timestamp", strconv.FormatInt(record.Timestamp, 10))
	values.Add("mac", record.Mac)
	values.Add("sequence_id", strconv.Itoa(record.SequenceId))
	values.Add("rssi", strconv.Itoa(record.Rssi))
	values.Add("ap_name", record.ApName)
	return
}

func (r ProbeRecord) String() string {
	ltime := time.Unix(r.Timestamp, 0).Local()
	return fmt.Sprintf("%s,%d,%s,%d,%d,%s", ltime.Format(time.RFC3339), r.Timestamp, r.Mac, r.SequenceId, r.Rssi, r.ApName)
}
