package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/class_manager/pkg/utils"
)

const (
	holidayAPIURL = "https://www.shuyz.com/githubfiles/china-holiday-calender/master/holidayAPI.json"
	cacheDuration = 90 * 24 * time.Hour // 90天缓存
)

// HolidayInfo 单个节假日信息
type HolidayInfo struct {
	Name      string   `json:"Name"`
	StartDate string   `json:"StartDate"`
	EndDate   string   `json:"EndDate"`
	Duration  int      `json:"Duration"`
	CompDays  []string `json:"CompDays"`
}

// holidayAPIResponse API响应结构
type holidayAPIResponse struct {
	Years map[string][]HolidayInfo `json:"Years"`
}

// DateHolidayInfo 某天的节假日信息
type DateHolidayInfo struct {
	IsHoliday   bool   `json:"is_holiday"`
	IsCompDay   bool   `json:"is_comp_day"` // 调休上班日
	HolidayName string `json:"holiday_name"`
}

// HolidayService 节假日服务
type HolidayService struct {
	mu         sync.RWMutex
	cache      *holidayAPIResponse
	cacheTime  time.Time
	fetching   bool
	httpClient *http.Client
}

var (
	holidayServiceInstance *HolidayService
	holidayOnce            sync.Once
)

// GetHolidayService 获取节假日服务单例
func GetHolidayService() *HolidayService {
	holidayOnce.Do(func() {
		holidayServiceInstance = &HolidayService{
			httpClient: &http.Client{Timeout: 5 * time.Second},
		}
	})
	return holidayServiceInstance
}

// fetchFromAPI 从远程API获取节假日数据
func (s *HolidayService) fetchFromAPI() (*holidayAPIResponse, error) {
	resp, err := s.httpClient.Get(holidayAPIURL)
	if err != nil {
		return nil, fmt.Errorf("请求节假日API失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("节假日API返回状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取节假日API响应失败: %w", err)
	}

	var result holidayAPIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析节假日API响应失败: %w", err)
	}

	return &result, nil
}

// backgroundFetch 在后台异步获取节假日数据
func (s *HolidayService) backgroundFetch() {
	defer func() {
		if rv := recover(); rv != nil {
			utils.LogErrorf("backgroundFetch panic: %v", rv)
		}
		s.mu.Lock()
		s.fetching = false
		s.mu.Unlock()
	}()

	utils.LogInfof("正在后台获取节假日数据...")
	data, err := s.fetchFromAPI()
	if err != nil {
		utils.LogErrorf("获取节假日数据失败: %v", err)
		return
	}

	// 使用匿名函数 + defer Unlock 确保锁总是被释放，避免 panic 时死锁
	func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.cache = data
		s.cacheTime = time.Now()
	}()

	utils.LogInfof("节假日数据获取成功，共 %d 个年份", len(data.Years))
}

// getCachedData 获取缓存的数据，如果缓存为空则触发后台异步获取
// 此方法不会阻塞，始终立即返回
func (s *HolidayService) getCachedData() *holidayAPIResponse {
	s.mu.RLock()
	if s.cache != nil && time.Since(s.cacheTime) < cacheDuration {
		data := s.cache
		s.mu.RUnlock()
		return data
	}
	s.mu.RUnlock()

	// 缓存为空或过期，触发后台异步获取（不阻塞）
	s.mu.Lock()
	if !s.fetching {
		s.fetching = true
		go s.backgroundFetch()
	}
	data := s.cache // 返回旧缓存（可能为nil）
	s.mu.Unlock()
	return data
}

// GetMonthHolidays 获取指定月份的节假日信息
// month 格式: "2026-07"
// 返回: map[dateStr]DateHolidayInfo
func (s *HolidayService) GetMonthHolidays(month string) (map[string]DateHolidayInfo, error) {
	data := s.getCachedData()
	if data == nil {
		// 缓存尚不可用，返回空map（后台正在获取，下次请求会有数据）
		return map[string]DateHolidayInfo{}, nil
	}

	// 解析月份
	parts := strings.Split(month, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("月份格式错误，应为 YYYY-MM: %s", month)
	}
	year := parts[0]

	yearHolidays, ok := data.Years[year]
	if !ok {
		// 该年份没有节假日数据，返回空map
		return map[string]DateHolidayInfo{}, nil
	}

	result := map[string]DateHolidayInfo{}

	// 遍历该年的所有节假日，展开日期范围
	for _, h := range yearHolidays {
		// 展开假期日期范围
		start, err := time.Parse("2006-01-02", h.StartDate)
		if err != nil {
			continue
		}
		end, err := time.Parse("2006-01-02", h.EndDate)
		if err != nil {
			continue
		}

		for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
			dateStr := d.Format("2006-01-02")
			if strings.HasPrefix(dateStr, month) {
				result[dateStr] = DateHolidayInfo{
					IsHoliday:   true,
					HolidayName: h.Name,
				}
			}
		}

		// 标记调休上班日
		for _, compDay := range h.CompDays {
			if strings.HasPrefix(compDay, month) {
				result[compDay] = DateHolidayInfo{
					IsCompDay:   true,
					HolidayName: h.Name,
				}
			}
		}
	}

	return result, nil
}
