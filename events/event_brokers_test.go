package events

// Wire-contract tests for the replay/broker half of the events service:
// Seek, SetEventSynchronizationPoint, and the event-broker CRUD.

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSeekAndSynchronizationPoint(t *testing.T) {
	t.Run("Seek", func(t *testing.T) {
		caller := newFakeCaller(func(action, reqXML string) (string, error) {
			if !strings.Contains(action, "Seek") {
				return "", errors.New("unexpected action " + action)
			}

			if !strings.Contains(reqXML, "2026-01-02T03:04:05Z") {
				return "", errors.New("UtcTime missing from request")
			}

			return `<SeekResponse/>`, nil
		})

		when := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		if err := New(caller).Seek(context.Background(), "sub-1", when, false); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Seek requires reference", func(t *testing.T) {
		if err := New(newFakeCaller(nil)).Seek(context.Background(), "", time.Now(), false); !errors.Is(err, ErrInvalidSubscriptionReference) {
			t.Fatalf("err = %v, want ErrInvalidSubscriptionReference", err)
		}
	})

	t.Run("SetEventSynchronizationPoint", func(t *testing.T) {
		caller := newFakeCaller(func(action, _ string) (string, error) {
			if !strings.Contains(action, "SetSynchronizationPoint") {
				return "", errors.New("unexpected action " + action)
			}

			return `<SetSynchronizationPointResponse/>`, nil
		})

		if err := New(caller).SetEventSynchronizationPoint(context.Background(), "sub-1"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("SetEventSynchronizationPoint requires reference", func(t *testing.T) {
		err := New(newFakeCaller(nil)).SetEventSynchronizationPoint(context.Background(), "")
		if !errors.Is(err, ErrInvalidSubscriptionReference) {
			t.Fatalf("err = %v, want ErrInvalidSubscriptionReference", err)
		}
	})
}

func TestEventBrokerCRUD(t *testing.T) {
	t.Run("AddEventBroker", func(t *testing.T) {
		caller := newFakeCaller(func(action, reqXML string) (string, error) {
			if !strings.Contains(action, "AddEventBroker") {
				return "", errors.New("unexpected action " + action)
			}

			if !strings.Contains(reqXML, "mqtt://broker.example.org") {
				return "", errors.New("broker address missing from request")
			}

			return `<AddEventBrokerResponse/>`, nil
		})

		err := New(caller).AddEventBroker(context.Background(), &EventBrokerConfig{
			Address:     "mqtt://broker.example.org",
			TopicPrefix: "onvif",
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("AddEventBroker validation", func(t *testing.T) {
		svc := New(newFakeCaller(nil))

		if err := svc.AddEventBroker(context.Background(), nil); !errors.Is(err, ErrEventBrokerConfigNil) {
			t.Errorf("nil config: err = %v", err)
		}

		err := svc.AddEventBroker(context.Background(), &EventBrokerConfig{})
		if !errors.Is(err, ErrInvalidEventBrokerAddress) {
			t.Errorf("empty address: err = %v", err)
		}
	})

	t.Run("DeleteEventBroker", func(t *testing.T) {
		caller := newFakeCaller(func(action, _ string) (string, error) {
			if !strings.Contains(action, "DeleteEventBroker") {
				return "", errors.New("unexpected action " + action)
			}

			return `<DeleteEventBrokerResponse/>`, nil
		})

		if err := New(caller).DeleteEventBroker(context.Background(), "mqtt://broker.example.org"); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("DeleteEventBroker requires address", func(t *testing.T) {
		if err := New(newFakeCaller(nil)).DeleteEventBroker(context.Background(), ""); !errors.Is(err, ErrInvalidEventBrokerAddress) {
			t.Fatalf("err = %v, want ErrInvalidEventBrokerAddress", err)
		}
	})

	t.Run("GetEventBrokers", func(t *testing.T) {
		caller := newFakeCaller(func(action, _ string) (string, error) {
			if !strings.Contains(action, "GetEventBrokers") {
				return "", errors.New("unexpected action " + action)
			}

			return `<GetEventBrokersResponse><EventBroker><Address>mqtt://a</Address><TopicPrefix>t</TopicPrefix><Status>Connected</Status></EventBroker></GetEventBrokersResponse>`, nil
		})

		brokers, err := New(caller).GetEventBrokers(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		if len(brokers) != 1 || brokers[0].Address != "mqtt://a" || brokers[0].Status != "Connected" {
			t.Fatalf("brokers = %+v", brokers)
		}
	})
}
