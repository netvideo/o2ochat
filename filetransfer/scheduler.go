package filetransfer

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

type SchedulerImpl struct {
	mu                sync.RWMutex
	chunkAvailability map[string]map[int][]string
	chunkStatus       map[string]map[int]chunkStatus
	schedulerStats    map[string]*SchedulerStats
	bandwidthPeers    map[string]float64
	retryCount        map[string]map[int]int
	maxRetries        int
}

type chunkStatus struct {
	inProgress bool
	completed  bool
	failed     bool
	lastPeer   string
}

func NewScheduler() Scheduler {
	return &SchedulerImpl{
		chunkAvailability: make(map[string]map[int][]string),
		chunkStatus:       make(map[string]map[int]chunkStatus),
		schedulerStats:    make(map[string]*SchedulerStats),
		bandwidthPeers:    make(map[string]float64),
		retryCount:        make(map[string]map[int]int),
		maxRetries:        3,
	}
}

func NewSchedulerImpl() Scheduler {
	return NewScheduler()
}

func (s *SchedulerImpl) SelectNextChunk(fileID string, availableChunks map[int][]string) (int, []string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(availableChunks) == 0 {
		return -1, nil, ErrInsufficientPeers
	}

	s.chunkAvailability[fileID] = availableChunks
	if s.chunkStatus[fileID] == nil {
		s.chunkStatus[fileID] = make(map[int]chunkStatus)
	}
	if s.retryCount[fileID] == nil {
		s.retryCount[fileID] = make(map[int]int)
	}

	var rarestChunk int
	minAvailability := math.MaxInt32
	availableList := make([]int, 0, len(availableChunks))

	for chunkIndex, peers := range availableChunks {
		status := s.chunkStatus[fileID][chunkIndex]
		if status.completed || status.inProgress {
			continue
		}

		availability := len(peers)
		if availability < minAvailability {
			minAvailability = availability
			rarestChunk = chunkIndex
			availableList = []int{chunkIndex}
		} else if availability == minAvailability {
			availableList = append(availableList, chunkIndex)
		}
	}

	if len(availableList) == 0 {
		return -1, nil, nil
	}

	if len(availableList) > 1 {
		rarestChunk = availableList[rand.Intn(len(availableList))]
	}

	s.chunkStatus[fileID][rarestChunk] = chunkStatus{
		inProgress: true,
		completed:  false,
		failed:     false,
	}

	if s.schedulerStats[fileID] == nil {
		s.schedulerStats[fileID] = &SchedulerStats{
			ChunkAvailability: make(map[int]int),
		}
	}
	s.schedulerStats[fileID].TotalScheduled++
	s.schedulerStats[fileID].Pending = len(availableList)

	return rarestChunk, availableChunks[rarestChunk], nil
}

func (s *SchedulerImpl) AssignDownloadTasks(fileID string, maxConcurrent int) (map[int][]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.chunkAvailability[fileID] == nil {
		return nil, ErrFileNotFound
	}

	assignments := make(map[int][]string)
	count := 0

	for chunkIndex, peers := range s.chunkAvailability[fileID] {
		status := s.chunkStatus[fileID][chunkIndex]
		if status.completed || status.inProgress {
			continue
		}

		if count >= maxConcurrent {
			break
		}

		selectedPeers := s.selectPeersByBandwidth(peers)
		assignments[chunkIndex] = selectedPeers

		s.chunkStatus[fileID][chunkIndex] = chunkStatus{
			inProgress: true,
			completed:  false,
			failed:     false,
		}
		count++
	}

	return assignments, nil
}

func (s *SchedulerImpl) UpdateChunkStatus(fileID string, index int, success bool, peerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.chunkStatus[fileID] == nil {
		s.chunkStatus[fileID] = make(map[int]chunkStatus)
	}

	currentStatus := s.chunkStatus[fileID][index]

	if success {
		currentStatus.completed = true
		currentStatus.inProgress = false
		currentStatus.failed = false

		if s.schedulerStats[fileID] != nil {
			s.schedulerStats[fileID].Successful++
			s.schedulerStats[fileID].Pending--
		}

		if s.bandwidthPeers[peerID] > 0 {
			s.bandwidthPeers[peerID] = s.bandwidthPeers[peerID] * 1.05
		}
	} else {
		currentStatus.inProgress = false
		currentStatus.failed = true
		currentStatus.lastPeer = peerID

		if s.retryCount[fileID] == nil {
			s.retryCount[fileID] = make(map[int]int)
		}
		s.retryCount[fileID][index]++

		if s.schedulerStats[fileID] != nil {
			s.schedulerStats[fileID].Failed++
			s.schedulerStats[fileID].Pending--
		}

		if s.bandwidthPeers[peerID] > 0 {
			s.bandwidthPeers[peerID] = s.bandwidthPeers[peerID] * 0.8
		}

		if s.retryCount[fileID][index] < s.maxRetries {
			currentStatus.inProgress = false
			currentStatus.failed = false
		}
	}

	s.chunkStatus[fileID][index] = currentStatus

	return nil
}

func (s *SchedulerImpl) GetSchedulerStats(fileID string) (*SchedulerStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats, ok := s.schedulerStats[fileID]
	if !ok {
		return &SchedulerStats{
			ChunkAvailability: make(map[int]int),
		}, nil
	}

	result := &SchedulerStats{
		TotalScheduled:    stats.TotalScheduled,
		Successful:        stats.Successful,
		Failed:            stats.Failed,
		Pending:           stats.Pending,
		ChunkAvailability: make(map[int]int),
	}

	for k, v := range stats.ChunkAvailability {
		result.ChunkAvailability[k] = v
	}

	return result, nil
}

func (s *SchedulerImpl) selectPeersByBandwidth(peers []string) []string {
	if len(peers) == 0 {
		return peers
	}

	type peerBandwidth struct {
		peerID    string
		bandwidth float64
	}

	peerList := make([]peerBandwidth, len(peers))
	for i, peerID := range peers {
		bandwidth := s.bandwidthPeers[peerID]
		if bandwidth == 0 {
			bandwidth = 100.0 + rand.Float64()*50
			s.bandwidthPeers[peerID] = bandwidth
		}
		peerList[i] = peerBandwidth{peerID, bandwidth}
	}

	for i := 0; i < len(peerList)-1; i++ {
		for j := i + 1; j < len(peerList); j++ {
			if peerList[i].bandwidth < peerList[j].bandwidth {
				peerList[i], peerList[j] = peerList[j], peerList[i]
			}
		}
	}

	result := make([]string, len(peerList))
	for i, pb := range peerList {
		result[i] = pb.peerID
	}

	return result
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
