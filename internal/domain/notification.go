package domain

type Notification struct {
	id               NotificationID
	userID           UserID
	isAllEnabled     bool
	isYoyoEnabled    bool
	isReportEnabled  bool
	isMessageEnabled bool
	subscription     string
}

func NewNotification(
	id NotificationID,
	userID UserID,
	isAllEnabled bool,
	isYoyoEnabled bool,
	isReportEnabled bool,
	isMessageEnabled bool,
	subscription string,
) Notification {
	return Notification{
		id:               id,
		userID:           userID,
		isAllEnabled:     isAllEnabled,
		isYoyoEnabled:    isYoyoEnabled,
		isReportEnabled:  isReportEnabled,
		isMessageEnabled: isMessageEnabled,
		subscription:     subscription,
	}
}

func (n Notification) ID() NotificationID {
	return n.id
}

func (n Notification) UserID() UserID {
	return n.userID
}

func (n Notification) IsAllEnabled() bool {
	return n.isAllEnabled
}

func (n Notification) IsYoyoEnabled() bool {
	return n.isYoyoEnabled
}

func (n Notification) IsReportEnabled() bool {
	return n.isReportEnabled
}

func (n Notification) IsMessageEnabled() bool {
	return n.isMessageEnabled
}

func (n Notification) Subscription() string {
	return n.subscription
}

type NotificationRepository interface {
	Create(notification Notification) (Notification, error)
	Update(notification Notification) (Notification, error)
	FindByUserID(userID UserID) (Notification, error)
}
