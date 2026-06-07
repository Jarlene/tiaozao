package service

import (
	"testing"

	"flea-market/internal/model"
)

func TestOrderFSM_LegalTransitions(t *testing.T) {
	fsm := NewOrderFSM()

	tests := []struct {
		name    string
		current model.OrderStatus
		event   OrderEvent
		want    model.OrderStatus
	}{
		// 待付款 → 已取消
		{"pending_payment→cancelled", model.OrderStatusPendingPayment, OrderEventCancel, model.OrderStatusCancelled},
		// 待付款 → 已付款
		{"pending_payment→paid", model.OrderStatusPendingPayment, OrderEventPay, model.OrderStatusPaid},
		// 已付款 → 待发货
		{"paid→pending_shipment", model.OrderStatusPaid, OrderEventNotifySeller, model.OrderStatusPendingShipment},
		// 待发货 → 已发货
		{"pending_shipment→shipped", model.OrderStatusPendingShipment, OrderEventShip, model.OrderStatusShipped},
		// 已发货 → 已收货
		{"shipped→received", model.OrderStatusShipped, OrderEventConfirmReceive, model.OrderStatusReceived},
		// 已发货 → 退款中
		{"shipped→refunding", model.OrderStatusShipped, OrderEventRequestRefund, model.OrderStatusRefunding},
		// 已收货 → 已完成
		{"received→completed", model.OrderStatusReceived, OrderEventComplete, model.OrderStatusCompleted},
		// 已收货 → 退款中
		{"received→refunding", model.OrderStatusReceived, OrderEventRequestRefund, model.OrderStatusRefunding},
		// 退款中 → 已完成
		{"refunding→completed", model.OrderStatusRefunding, OrderEventRefundSuccess, model.OrderStatusCompleted},
		// 退款中 → 纠纷处理
		{"refunding→dispute", model.OrderStatusRefunding, OrderEventRaiseDispute, model.OrderStatusDispute},
		// 纠纷处理 → 已完成
		{"dispute→completed", model.OrderStatusDispute, OrderEventArbitrate, model.OrderStatusCompleted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fsm.Transition(tt.current, tt.event)
			if err != nil {
				t.Fatalf("unexpected error for transition %s: %v", tt.name, err)
			}
			if got != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestOrderFSM_IllegalTransitions(t *testing.T) {
	fsm := NewOrderFSM()

	// 每个状态与它不应支持的事件组合
	illegalCases := []struct {
		name    string
		current model.OrderStatus
		event   OrderEvent
	}{
		// 待付款不能直接完成
		{"pending_payment→completed(invalid)", model.OrderStatusPendingPayment, OrderEventComplete},
		// 待付款不能发货
		{"pending_payment→ship(invalid)", model.OrderStatusPendingPayment, OrderEventShip},
		// 待付款不能确认收货
		{"pending_payment→confirm_receive(invalid)", model.OrderStatusPendingPayment, OrderEventConfirmReceive},
		// 待付款不能退款
		{"pending_payment→request_refund(invalid)", model.OrderStatusPendingPayment, OrderEventRequestRefund},
		// 已取消不能做任何事
		{"cancelled→pay(invalid)", model.OrderStatusCancelled, OrderEventPay},
		{"cancelled→cancel(invalid)", model.OrderStatusCancelled, OrderEventCancel},
		// 已付款不能取消（资金已托管）
		{"paid→cancel(invalid)", model.OrderStatusPaid, OrderEventCancel},
		// 已付款不能直接发货（必须先通知卖家）
		{"paid→ship(invalid)", model.OrderStatusPaid, OrderEventShip},
		// 待发货不能付款
		{"pending_shipment→pay(invalid)", model.OrderStatusPendingShipment, OrderEventPay},
		// 已完成不能退款
		{"completed→request_refund(invalid)", model.OrderStatusCompleted, OrderEventRequestRefund},
		{"completed→refund_success(invalid)", model.OrderStatusCompleted, OrderEventRefundSuccess},
		// 纠纷处理不能退款成功（只能仲裁）
		{"dispute→refund_success(invalid)", model.OrderStatusDispute, OrderEventRefundSuccess},
		// 纠纷处理不能发起纠纷
		{"dispute→raise_dispute(invalid)", model.OrderStatusDispute, OrderEventRaiseDispute},
	}

	for _, tc := range illegalCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := fsm.Transition(tc.current, tc.event)
			if err == nil {
				t.Fatalf("expected error for illegal transition, got nil")
			}
		})
	}
}

func TestOrderFSM_Can(t *testing.T) {
	fsm := NewOrderFSM()

	// 合法转换应该返回 true
	if !fsm.Can(model.OrderStatusPendingPayment, OrderEventPay) {
		t.Error("expected Can(pending_payment, pay) = true")
	}
	if !fsm.Can(model.OrderStatusPendingPayment, OrderEventCancel) {
		t.Error("expected Can(pending_payment, cancel) = true")
	}

	// 非法转换应该返回 false
	if fsm.Can(model.OrderStatusPendingPayment, OrderEventShip) {
		t.Error("expected Can(pending_payment, ship) = false")
	}
	if fsm.Can(model.OrderStatusCompleted, OrderEventRequestRefund) {
		t.Error("expected Can(completed, request_refund) = false")
	}
	if fsm.Can(model.OrderStatusCancelled, OrderEventPay) {
		t.Error("expected Can(cancelled, pay) = false")
	}
}

func TestOrderFSM_ValidEvents(t *testing.T) {
	fsm := NewOrderFSM()

	// 待付款应该有 2 个合法事件：取消和付款
	events := fsm.ValidEvents(model.OrderStatusPendingPayment)
	if len(events) != 2 {
		t.Fatalf("expected 2 valid events for pending_payment, got %d", len(events))
	}

	hasCancel, hasPay := false, false
	for _, e := range events {
		if e == OrderEventCancel {
			hasCancel = true
		}
		if e == OrderEventPay {
			hasPay = true
		}
	}
	if !hasCancel || !hasPay {
		t.Fatal("pending_payment should have cancel and pay events")
	}

	// 已取消应该没有合法事件
	events = fsm.ValidEvents(model.OrderStatusCancelled)
	if len(events) != 0 {
		t.Fatalf("expected 0 valid events for cancelled, got %d", len(events))
	}

	// 已发货应该有 2 个合法事件（收货和退款）
	events = fsm.ValidEvents(model.OrderStatusShipped)
	if len(events) != 2 {
		t.Fatalf("expected 2 valid events for shipped, got %d", len(events))
	}
}

func TestOrderFSM_ValidTransitions(t *testing.T) {
	fsm := NewOrderFSM()

	transitions := fsm.ValidTransitions(model.OrderStatusShipped)
	if len(transitions) != 2 {
		t.Fatalf("expected 2 valid transitions for shipped, got %d", len(transitions))
	}

	hasReceived, hasRefunding := false, false
	for _, s := range transitions {
		if s == model.OrderStatusReceived {
			hasReceived = true
		}
		if s == model.OrderStatusRefunding {
			hasRefunding = true
		}
	}
	if !hasReceived || !hasRefunding {
		t.Fatal("shipped should transition to received and refunding")
	}
}

func TestOrderFSM_PayThenAutoNotify(t *testing.T) {
	fsm := NewOrderFSM()

	// 付款
	status, err := fsm.Transition(model.OrderStatusPendingPayment, OrderEventPay)
	if err != nil {
		t.Fatalf("pay failed: %v", err)
	}
	if status != model.OrderStatusPaid {
		t.Fatalf("expected paid, got %d", status)
	}

	// 自动通知卖家
	status, err = fsm.Transition(status, OrderEventNotifySeller)
	if err != nil {
		t.Fatalf("notify_seller failed: %v", err)
	}
	if status != model.OrderStatusPendingShipment {
		t.Fatalf("expected pending_shipment, got %d", status)
	}
}

func TestOrderFSM_FullHappyPath(t *testing.T) {
	fsm := NewOrderFSM()

	events := []struct {
		event OrderEvent
		want  model.OrderStatus
	}{
		{OrderEventPay, model.OrderStatusPaid},
		{OrderEventNotifySeller, model.OrderStatusPendingShipment},
		{OrderEventShip, model.OrderStatusShipped},
		{OrderEventConfirmReceive, model.OrderStatusReceived},
		{OrderEventComplete, model.OrderStatusCompleted},
	}

	current := model.OrderStatusPendingPayment
	for _, step := range events {
		next, err := fsm.Transition(current, step.event)
		if err != nil {
			t.Fatalf("transition %s from %d failed: %v", step.event, current, err)
		}
		if next != step.want {
			t.Fatalf("from %d via %s: expected %d, got %d", current, step.event, step.want, next)
		}
		current = next
	}
}

func TestOrderFSM_RefundThenDisputeThenArbitrate(t *testing.T) {
	fsm := NewOrderFSM()

	// 正常流程到已发货
	current := model.OrderStatusPendingPayment
	current, _ = fsm.Transition(current, OrderEventPay)
	current, _ = fsm.Transition(current, OrderEventNotifySeller)
	current, _ = fsm.Transition(current, OrderEventShip)

	// 申请退款
	current, err := fsm.Transition(current, OrderEventRequestRefund)
	if err != nil {
		t.Fatalf("request_refund failed: %v", err)
	}
	if current != model.OrderStatusRefunding {
		t.Fatalf("expected refunding, got %d", current)
	}

	// 发起纠纷
	current, err = fsm.Transition(current, OrderEventRaiseDispute)
	if err != nil {
		t.Fatalf("raise_dispute failed: %v", err)
	}
	if current != model.OrderStatusDispute {
		t.Fatalf("expected dispute, got %d", current)
	}

	// 管理员仲裁
	current, err = fsm.Transition(current, OrderEventArbitrate)
	if err != nil {
		t.Fatalf("arbitrate failed: %v", err)
	}
	if current != model.OrderStatusCompleted {
		t.Fatalf("expected completed, got %d", current)
	}
}

func TestOrderFSM_RefundFromReceived(t *testing.T) {
	fsm := NewOrderFSM()

	current := model.OrderStatusPendingPayment
	current, _ = fsm.Transition(current, OrderEventPay)
	current, _ = fsm.Transition(current, OrderEventNotifySeller)
	current, _ = fsm.Transition(current, OrderEventShip)
	current, _ = fsm.Transition(current, OrderEventConfirmReceive)

	// 已收货状态申请退款
	current, err := fsm.Transition(current, OrderEventRequestRefund)
	if err != nil {
		t.Fatalf("request_refund from received failed: %v", err)
	}
	if current != model.OrderStatusRefunding {
		t.Fatalf("expected refunding, got %d", current)
	}

	// 退款成功
	current, err = fsm.Transition(current, OrderEventRefundSuccess)
	if err != nil {
		t.Fatalf("refund_success failed: %v", err)
	}
	if current != model.OrderStatusCompleted {
		t.Fatalf("expected completed, got %d", current)
	}
}
