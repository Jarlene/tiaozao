package service

import (
	"fmt"

	"flea-market/internal/model"
)

// OrderEvent 触发状态变更的事件
type OrderEvent string

const (
	OrderEventCancel         OrderEvent = "cancel"          // 买家取消
	OrderEventPay            OrderEvent = "pay"             // 买家付款
	OrderEventNotifySeller   OrderEvent = "notify_seller"   // 通知卖家（系统自动）
	OrderEventShip           OrderEvent = "ship"            // 卖家发货
	OrderEventConfirmReceive OrderEvent = "confirm_receive" // 买家确认收货
	OrderEventComplete       OrderEvent = "complete"        // 完成
	OrderEventRequestRefund  OrderEvent = "request_refund"  // 买家申请退款
	OrderEventRefundSuccess  OrderEvent = "refund_success"  // 退款成功（系统/卖家同意）
	OrderEventApproveRefund  OrderEvent = "approve_refund"  // 卖家同意退款（日志用）
	OrderEventRejectRefund   OrderEvent = "reject_refund"   // 卖家拒绝退款（日志用）
	OrderEventRaiseDispute   OrderEvent = "raise_dispute"   // 升级为纠纷
	OrderEventArbitrate      OrderEvent = "arbitrate"       // 管理员仲裁
)

// OrderEventNames 事件中文名（用于操作日志）
var OrderEventNames = map[OrderEvent]string{
	OrderEventCancel:         "取消订单",
	OrderEventPay:            "付款",
	OrderEventNotifySeller:   "通知卖家",
	OrderEventShip:           "发货",
	OrderEventConfirmReceive: "确认收货",
	OrderEventComplete:       "完成订单",
	OrderEventRequestRefund:  "申请退款",
	OrderEventRefundSuccess:  "退款成功",
	OrderEventApproveRefund:  "同意退款",
	OrderEventRejectRefund:   "拒绝退款",
	OrderEventRaiseDispute:   "发起纠纷",
	OrderEventArbitrate:      "仲裁",
}

// transitionRule 定义一条转换规则
type transitionRule struct {
	From model.OrderStatus
	To   model.OrderStatus
}

// OrderFSM 订单状态机（显式模式）
// 通过 transitionTable 明确定义所有合法转换路径
type OrderFSM struct {
	// transitionMap: current -> {event -> target}
	transitionMap map[model.OrderStatus]map[OrderEvent]model.OrderStatus
}

// NewOrderFSM 创建状态机并初始化转换表
func NewOrderFSM() *OrderFSM {
	fsm := &OrderFSM{
		transitionMap: make(map[model.OrderStatus]map[OrderEvent]model.OrderStatus),
	}
	fsm.initTransitions()
	return fsm
}

// initTransitions 初始化完整的状态转换表
// 覆盖 STORY-5.1.2 定义的所有合法路径
func (fsm *OrderFSM) initTransitions() {
	rules := []transitionRule{
		{From: model.OrderStatusPendingPayment, To: model.OrderStatusCancelled},
		{From: model.OrderStatusPendingPayment, To: model.OrderStatusPaid},
		{From: model.OrderStatusPaid, To: model.OrderStatusPendingShipment},
		{From: model.OrderStatusPendingShipment, To: model.OrderStatusShipped},
		{From: model.OrderStatusShipped, To: model.OrderStatusReceived},
		{From: model.OrderStatusShipped, To: model.OrderStatusRefunding},
		{From: model.OrderStatusReceived, To: model.OrderStatusCompleted},
		{From: model.OrderStatusReceived, To: model.OrderStatusRefunding},
		{From: model.OrderStatusRefunding, To: model.OrderStatusCompleted},
		{From: model.OrderStatusRefunding, To: model.OrderStatusDispute},
		{From: model.OrderStatusDispute, To: model.OrderStatusCompleted},
	}

	// 事件映射
	// 注：FSM 用 transitionRule（From→To 对）作为 map key，同一对只对应一个事件
	// 语义区分（如"卖家同意"vs"系统退款"）由 service 方法在状态日志中体现
	eventMap := map[transitionRule]OrderEvent{
		{From: model.OrderStatusPendingPayment, To: model.OrderStatusCancelled}:       OrderEventCancel,
		{From: model.OrderStatusPendingPayment, To: model.OrderStatusPaid}:            OrderEventPay,
		{From: model.OrderStatusPaid, To: model.OrderStatusPendingShipment}:           OrderEventNotifySeller,
		{From: model.OrderStatusPendingShipment, To: model.OrderStatusShipped}:        OrderEventShip,
		{From: model.OrderStatusShipped, To: model.OrderStatusReceived}:               OrderEventConfirmReceive,
		{From: model.OrderStatusShipped, To: model.OrderStatusRefunding}:              OrderEventRequestRefund,
		{From: model.OrderStatusReceived, To: model.OrderStatusCompleted}:             OrderEventComplete,
		{From: model.OrderStatusReceived, To: model.OrderStatusRefunding}:             OrderEventRequestRefund,
		{From: model.OrderStatusRefunding, To: model.OrderStatusCompleted}:            OrderEventRefundSuccess,
		{From: model.OrderStatusRefunding, To: model.OrderStatusDispute}:              OrderEventRaiseDispute,
		{From: model.OrderStatusDispute, To: model.OrderStatusCompleted}:              OrderEventArbitrate,
	}

	for _, rule := range rules {
		event := eventMap[rule]
		if fsm.transitionMap[rule.From] == nil {
			fsm.transitionMap[rule.From] = make(map[OrderEvent]model.OrderStatus)
		}
		fsm.transitionMap[rule.From][event] = rule.To
	}
}

// Transition 根据当前状态和事件返回目标状态
// 如果转换非法，返回错误
func (fsm *OrderFSM) Transition(current model.OrderStatus, event OrderEvent) (model.OrderStatus, error) {
	events, ok := fsm.transitionMap[current]
	if !ok {
		return current, fmt.Errorf("状态 %d 没有定义任何转换", current)
	}

	target, ok := events[event]
	if !ok {
		return current, fmt.Errorf("非法状态转换：从 %s 通过 %s", model.OrderStatusNames[current], OrderEventNames[event])
	}

	return target, nil
}

// Can 检查从当前状态通过指定事件是否可以转换
func (fsm *OrderFSM) Can(current model.OrderStatus, event OrderEvent) bool {
	_, err := fsm.Transition(current, event)
	return err == nil
}

// ValidEvents 返回当前状态下所有合法事件
func (fsm *OrderFSM) ValidEvents(current model.OrderStatus) []OrderEvent {
	events, ok := fsm.transitionMap[current]
	if !ok {
		return nil
	}
	result := make([]OrderEvent, 0, len(events))
	for event := range events {
		result = append(result, event)
	}
	return result
}

// ValidTransitions 返回当前状态下所有合法目标状态
func (fsm *OrderFSM) ValidTransitions(current model.OrderStatus) []model.OrderStatus {
	events, ok := fsm.transitionMap[current]
	if !ok {
		return nil
	}
	seen := make(map[model.OrderStatus]struct{})
	result := make([]model.OrderStatus, 0, len(events))
	for _, target := range events {
		if _, ok := seen[target]; !ok {
			seen[target] = struct{}{}
			result = append(result, target)
		}
	}
	return result
}
