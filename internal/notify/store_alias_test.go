package notify

import (
	"testing"
	"time"
)

// TestCreate_ScheduleAtIndependentFromInput 是针对 ScheduleAt 别名穿透 Bug 的回归测试。
//
// 现象：创建通知成功后，调用方修改原始 CreateInput 中的 ScheduleAt 值，
// 会导致服务内部存储的通知的计划发送时间也跟着改变（外部输入“穿透”到内部存储）。
//
// 预期：创建成功后，服务内部的通知数据应与调用方输入完全独立，
// 修改原始输入不应影响已存储的通知，也不应影响创建接口返回的快照。
//
// 该用例不依赖任何外部命令或网络，直接作用于 notify.Store，便于稳定复现。
func TestCreate_ScheduleAtIndependentFromInput(t *testing.T) {
	s := New()
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)

	// 固定的回归输入：计划发送时间晚于 now，校验合法。
	expected := time.Date(2026, 8, 15, 13, 0, 0, 0, time.UTC)
	sched := expected
	in := CreateInput{
		ID:         "N1",
		Recipient:  "user-a",
		Content:    "你好",
		Priority:   PriorityNormal,
		ScheduleAt: &sched,
	}

	created, err := s.Create(in, now)
	if err != nil {
		t.Fatalf("Create 失败: %v", err)
	}

	// 调用方在创建成功后修改原始输入中的 ScheduleAt 值。
	*in.ScheduleAt = sched.Add(2 * time.Hour)

	// 断言 1：服务内部存储的计划发送时间不应被穿透修改。
	got, err := s.Get("N1")
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if got.ScheduleAt == nil {
		t.Fatal("内部存储的 ScheduleAt 为 nil")
	}
	if !got.ScheduleAt.Equal(expected) {
		t.Fatalf("内部存储的 ScheduleAt 被输入修改穿透: got %v, want %v",
			*got.ScheduleAt, expected)
	}

	// 断言 2：创建接口返回的快照同样不应被穿透修改。
	if created.ScheduleAt == nil {
		t.Fatal("创建返回的 ScheduleAt 为 nil")
	}
	if !created.ScheduleAt.Equal(expected) {
		t.Fatalf("创建返回的 ScheduleAt 被输入修改穿透: got %v, want %v",
			*created.ScheduleAt, expected)
	}
}
