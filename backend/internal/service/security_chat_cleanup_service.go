package service

import (
	"context"
	"log"
	"time"
)

const securityChatCleanupJobName = "security_chat_cleanup"

type SecurityChatCleanupService struct {
	chatService *SecurityChatService
	timingWheel *TimingWheelService
}

func NewSecurityChatCleanupService(chatService *SecurityChatService, timingWheel *TimingWheelService) *SecurityChatCleanupService {
	return &SecurityChatCleanupService{chatService: chatService, timingWheel: timingWheel}
}

func (s *SecurityChatCleanupService) Start() {
	if s == nil || s.chatService == nil || s.timingWheel == nil {
		return
	}
	interval := 1 * time.Hour
	s.timingWheel.ScheduleRecurring(securityChatCleanupJobName, interval, s.runOnce)
}

func (s *SecurityChatCleanupService) runOnce() {
	if s == nil || s.chatService == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	deleted, err := s.chatService.CleanupExpired(ctx)
	if err != nil {
		log.Printf("[SecurityChatCleanup] cleanup failed: %v", err)
		return
	}
	if deleted > 0 {
		log.Printf("[SecurityChatCleanup] deleted=%d", deleted)
	}
}
