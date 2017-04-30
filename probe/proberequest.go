package probe

import (
    "net/url"
    "strconv"
    "time"
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
