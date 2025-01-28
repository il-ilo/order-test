package main

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type Window struct {
	From time.Time
}

type AggregatedState struct {
	HaveAz bool
	HavePa bool

	StartTime time.Time `json:"StartTime,omitempty"`
	EndTime   time.Time `json:"EndTime,omitempty"`

	SrcIP                 string `json:"SrcIP,omitempty"`
	DestIP                string `json:"DestIP,omitempty"`
	Proto                 string `json:"Proto,omitempty"`
	Port                  int    `json:"Port"`
	TenantID              string `json:"IllumioTenantId,omitempty"`
	SourceMACAddress      string `json:"SourceMACAddress,omitempty"`
	DestinationMACAddress string `json:"DestinationMACAddress,omitempty"`

	SentBytes       int64  `json:"SentBytes,omitempty"`
	ReceivedBytes   int64  `json:"ReceivedBytes,omitempty"`
	PacketsSent     int64  `json:"PacketsSent,omitempty"`
	PacketsReceived int64  `json:"PacketsReceived,omitempty"`
	TrafficStatus   string `json:"TrafficStatus,omitempty"`
	FlowCount       int64  `json:"FlowCount,omitempty"`
}

func (s *AggregatedState) Apply(f Flow) {
	mismatch := false
	if f.DeviceProduct != "" {
		s.HavePa = true
	} else {
		s.HaveAz = true
	}
	st := parseTime(f.StartTime)
	if s.StartTime.IsZero() || s.StartTime.Before(st) {
		s.StartTime = st
	}
	et := parseTime(f.EndTime)
	if et.After(s.EndTime) {
		s.EndTime = et
	}

	if s.SourceMACAddress != f.SourceMACAddress {
		mismatch = true
		fmt.Printf("src MAC mismatch: %v != %v\n", s.SourceMACAddress, f.SourceMACAddress)
	}
	if s.DestinationMACAddress != f.DestinationMACAddress {
		mismatch = true
		fmt.Printf("dst MAC mismatch: %v != %v\n", s.DestinationMACAddress, f.DestinationMACAddress)
	}
	if s.TrafficStatus != f.TrafficStatus {
		mismatch = true
		fmt.Printf("status mismatch: %v != %v\n", s.TrafficStatus, f.TrafficStatus)
	}

	if mismatch {
		fmt.Printf("mismatch: %+v vs %+v\n", s, f)
	}

	s.SentBytes += f.SentBytes
	s.ReceivedBytes += f.ReceivedBytes
	s.PacketsSent += f.PacketsSent
	s.PacketsReceived += f.PacketsReceived
	s.FlowCount++
}

type State struct {
	Data   map[string]map[time.Time]*AggregatedState
	MaxCap int
}

func NewState(mx int) *State {
	return &State{
		Data:   make(map[string]map[time.Time]*AggregatedState),
		MaxCap: mx,
	}
}

func (s *State) Process(f Flow) {
	k := f.Key()
	valForKey, have := s.Data[k]
	if !have {
		if len(s.Data) == s.MaxCap {
			return
		}
		valForKey = make(map[time.Time]*AggregatedState)
		s.Data[k] = valForKey
	}
	st := parseTime(f.StartTime).UTC().Truncate(time.Hour)
	valForWin, have := valForKey[st]
	if !have {
		valForWin = &AggregatedState{
			HaveAz:                f.DeviceProduct == "",
			HavePa:                f.DeviceProduct != "",
			StartTime:             st,
			EndTime:               parseTime(f.EndTime),
			SrcIP:                 f.SrcIP,
			DestIP:                f.DestIP,
			Proto:                 f.Proto,
			Port:                  f.Port,
			TenantID:              f.TenantID,
			SourceMACAddress:      f.SourceMACAddress,
			DestinationMACAddress: f.DestinationMACAddress,
			SentBytes:             f.SentBytes,
			ReceivedBytes:         f.ReceivedBytes,
			PacketsSent:           f.PacketsSent,
			PacketsReceived:       f.PacketsReceived,
			TrafficStatus:         f.TrafficStatus,
			FlowCount:             1,
		}
		valForKey[st] = valForWin
	} else {
		valForWin.Apply(f)
	}
}

func (s *State) Describe(f io.Writer) {
	for k, v := range s.Data {
		fmt.Fprintln(f, "###########################################################")
		fmt.Fprintf(f, "%v:\n", k)
		for w, vv := range v {
			fmt.Fprintf(f, "%v to %v:\n", w.Format(time.RFC3339), w.Add(time.Hour).Format(time.RFC3339))
			o, err := json.MarshalIndent(vv, "  ", "  ")
			if err != nil {
				panic(err)
			}
			fmt.Fprintf(f, "%v\n", string(o))
		}
		fmt.Fprint(f, "###########################################################\n\n")
	}
}
