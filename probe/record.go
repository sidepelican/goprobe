package probe

import (
    "net/url"
    "strconv"
    "time"
    "fmt"
)

type ProbeRecord struct {
    Time       time.Time
    Mac        string
    SequenceId int
    Rssi       int
    ApName     string
}

func (record *ProbeRecord) Values() (values url.Values) {
    values.Add("time", record.Time.Local().String())
    values.Add("mac", record.Mac)
    values.Add("sequence_id", strconv.Itoa(record.SequenceId))
    values.Add("rssi", strconv.Itoa(record.Rssi))
    values.Add("ap_name", record.ApName)
    return
}

func (r ProbeRecord) String() string {
    return fmt.Sprintf("%s,%s,%d,%d,%s", r.Time.Format(time.RFC3339), r.Mac, r.SequenceId, r.Rssi, r.ApName)
}