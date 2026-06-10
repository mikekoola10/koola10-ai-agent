package swarms

import (
	"log"
	"sync"
	"koola10/pkg/vault"
)

type LedgerEvent struct {
	Description string
	Amount      float64
	Type        string
	Notes       string
}

type VaultKeeper struct {
	client   *vault.VaultClient
	eventCh  chan LedgerEvent
	wg       sync.WaitGroup
	shutdown chan struct{}
}

func NewVaultKeeper() *VaultKeeper {
	vk := &VaultKeeper{
		client:   vault.NewVaultClient(),
		eventCh:  make(chan LedgerEvent, 100),
		shutdown: make(chan struct{}),
	}
	vk.wg.Add(1)
	go vk.run()
	return vk
}

func (vk *VaultKeeper) run() {
	defer vk.wg.Done()
	log.Println("VaultKeeper swarm started")
	for {
		select {
		case event := <-vk.eventCh:
			err := vk.client.AddEntry(vault.VaultEntry{
				Description: event.Description,
				Amount:      event.Amount,
				Type:        event.Type,
				Notes:       event.Notes,
			})
			if err != nil {
				log.Printf("VaultKeeper error: %v", err)
				// Optionally retry after delay could be added here
			} else {
				log.Printf("Logged to vault: %s $%.2f", event.Description, event.Amount)
			}
		case <-vk.shutdown:
			log.Println("VaultKeeper swarm shutting down")
			return
		}
	}
}

func (vk *VaultKeeper) LogEvent(event LedgerEvent) {
	select {
	case vk.eventCh <- event:
	default:
		log.Println("VaultKeeper event channel full, dropping event")
	}
}

func (vk *VaultKeeper) Shutdown() {
	close(vk.shutdown)
	vk.wg.Wait()
}
