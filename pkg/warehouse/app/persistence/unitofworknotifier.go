package persistence

type unitOfWorkCompleteNotifier struct {
	ufw        UnitOfWork
	notifyFunc func()
}

func (ufw *unitOfWorkCompleteNotifier) Execute(lockName string, f func(p PersistentProvider) error) error {
	err := ufw.ufw.Execute(lockName, f)
	if err == nil {
		ufw.notifyFunc()
	}
	return err
}

func NewUnitOfWorkCompleteNotifier(ufw UnitOfWork, notifyFunc func()) UnitOfWork {
	return &unitOfWorkCompleteNotifier{ufw: ufw, notifyFunc: notifyFunc}
}
