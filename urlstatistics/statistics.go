// core from https://github.com/go-macaron/toolbox/blob/master/statistic.go
package urlstatistics

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/meilihao/water"
)

// Statistics struct
type Statistics struct {
	RequestUrl string
	RequestNum int64
	MinTime    time.Duration
	MaxTime    time.Duration
	TotalTime  time.Duration
	Codes      map[int]int
}

// UrlMap contains several statistics struct to log different data
type UrlMap struct {
	lock        sync.RWMutex
	LengthLimit int                               // limit the urlmap's length if it's equal to 0 there's no limit
	urlmap      map[string]map[string]*Statistics // urlmap[url][method]
}

type CodeCount struct {
	Code  int `json:"code"`
	Count int `json:"count"`
}

func (s *Statistics) CodesCount() []*CodeCount {
	ls := make([]*CodeCount, len(s.Codes))

	codes := make([]int, 0, len(s.Codes))
	for k := range s.Codes {
		codes = append(codes, k)
	}
	sort.Ints(codes)

	for i, v := range codes {
		ls[i] = &CodeCount{
			Code:  v,
			Count: s.Codes[v],
		}
	}

	return ls
}

func (s *Statistics) CodesCountString() string {
	ls := s.CodesCount()

	tmps := make([]string, len(ls))
	for i, v := range ls {
		tmps[i] = fmt.Sprintf("%d:%d", v.Code, v.Count)
	}

	return strings.Join(tmps, ", ")
}

var (
	m = &UrlMap{
		urlmap: make(map[string]map[string]*Statistics, 9),
	}
)

func addStatistics(requestMethod, requestUrl string, code int, requesttime time.Duration) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if code == 0 {
		code = http.StatusOK
	}

	if method, ok := m.urlmap[requestUrl]; ok {
		if s, ok := method[requestMethod]; ok {
			s.RequestNum += 1
			if s.MaxTime < requesttime {
				s.MaxTime = requesttime
			}
			if s.MinTime > requesttime {
				s.MinTime = requesttime
			}
			s.TotalTime += requesttime
			s.Codes[code] = s.Codes[code] + 1
		} else {
			nb := &Statistics{
				RequestUrl: requestUrl,
				RequestNum: 1,
				MinTime:    requesttime,
				MaxTime:    requesttime,
				TotalTime:  requesttime,
				Codes: map[int]int{
					code: 1,
				},
			}
			m.urlmap[requestUrl][requestMethod] = nb
		}

	} else {
		if m.LengthLimit > 0 && m.LengthLimit <= len(m.urlmap) {
			return
		}
		methodmap := make(map[string]*Statistics)
		nb := &Statistics{
			RequestUrl: requestUrl,
			RequestNum: 1,
			MinTime:    requesttime,
			MaxTime:    requesttime,
			TotalTime:  requesttime,
			Codes: map[int]int{
				code: 1,
			},
		}
		methodmap[requestMethod] = nb
		m.urlmap[requestUrl] = methodmap
	}
}

func URLStatistics() water.HandlerFunc {
	return func(c *water.Context) {
		start := time.Now()
		c.Next()

		addStatistics(c.Request.Method, c.Request.URL.Path, c.Status(), time.Since(start))
	}
}

func GetURLStatistics() *UrlMap {
	return m
}

// put url statistics result in io.Writer
func (m *UrlMap) GetMap(w io.Writer) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	sep := fmt.Sprintf("+%s+%s+%s+%s+%s+%s+%s+%s+\n",
		strings.Repeat("-", 51), strings.Repeat("-", 12), strings.Repeat("-", 18), strings.Repeat("-", 18),
		strings.Repeat("-", 18), strings.Repeat("-", 18), strings.Repeat("-", 18), strings.Repeat("-", 18))
	_, _ = fmt.Fprint(w, sep)
	_, _ = fmt.Fprintf(w, "| % -50s| % -10s | % -16s | % -16s | % -16s | % -16s | % -16s | % -16s |\n", "Request URL", "Method", "Times", "Status Times", "Total Used(s)", "Max Used(μs)", "Min Used(μs)", "Avg Used(μs)")
	_, _ = fmt.Fprint(w, sep)

	for k, v := range m.urlmap {
		for kk, vv := range v {
			fmt.Fprintf(w, "| % -50s| % -10s | % 16d | % -16s | % 16f | % 16.6f | % 16.6f | % 16.6f |\n",
				k, kk, vv.RequestNum, vv.CodesCountString(),
				vv.TotalTime.Seconds(), float64(vv.MaxTime.Nanoseconds())/1000, float64(vv.MinTime.Nanoseconds())/1000, float64(time.Duration(int64(vv.TotalTime)/vv.RequestNum).Nanoseconds())/1000,
			)
		}
	}
	_, _ = fmt.Fprint(w, sep)
}

type URLMapInfo struct {
	URL       string       `json:"url"`
	Method    string       `json:"method"`
	Times     int64        `json:"times"`
	TotalUsed float64      `json:"total_used"`
	MaxUsed   float64      `json:"max_used"`
	MinUsed   float64      `json:"min_used"`
	AvgUsed   float64      `json:"avg_used"`
	Codes     []*CodeCount `json:"codes"`
}

func (m *UrlMap) JSON(w io.Writer) {
	infos := make([]*URLMapInfo, 0, len(m.urlmap))
	for k, v := range m.urlmap {
		for kk, vv := range v {
			infos = append(infos, &URLMapInfo{
				URL:       k,
				Method:    kk,
				Times:     vv.RequestNum,
				TotalUsed: vv.TotalTime.Seconds(),
				MaxUsed:   float64(vv.MaxTime.Nanoseconds()) / 1000,
				MinUsed:   float64(vv.MinTime.Nanoseconds()) / 1000,
				AvgUsed:   float64(time.Duration(int64(vv.TotalTime)/vv.RequestNum).Nanoseconds()) / 1000,
				Codes:     vv.CodesCount(),
			})
		}
	}

	if err := json.NewEncoder(w).Encode(infos); err != nil {
		panic("URLMap.JSON: " + err.Error())
	}
}
