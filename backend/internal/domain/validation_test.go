package domain

import (
	"errors"
	"testing"
	"time"
)

func TestEnsureNoHTMLAllowsLegitimateNamesAndPunctuation(t *testing.T) {
	cases := []string{
		"O'Connor",
		"Jean-Luc Picard",
		"Dr. Mario & Luigi",
		"Renault / Nissan 2.0",
		"Cambio de aceite, filtro y bujias.",
	}
	for _, text := range cases {
		if err := EnsureNoHTML("test field", text); err != nil {
			t.Errorf("expected %q to be allowed, got error: %v", text, err)
		}
	}
}

func TestEnsureNoHTMLRejectsHTMLTags(t *testing.T) {
	cases := []string{
		"<script>alert(1)</script>",
		"O'Connor <script>alert(1)</script>",
		"<img src=x onerror=alert(1)>",
		"<div>Inyeccion</div>",
		"<b>Negrita</b>",
		"<a href='http://evil.com'>click</a>",
	}
	for _, text := range cases {
		err := EnsureNoHTML("test field", text)
		if err == nil {
			t.Errorf("expected %q to be rejected, but got nil", text)
		}
		if !errors.Is(err, ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	}
}

func TestCustomerRejectsHTML(t *testing.T) {
	_, err := NewCustomer("c-1", "O'Connor <script>alert(1)</script>", "12345678", "3001234567", "test@example.com", time.Now())
	if err == nil {
		t.Fatal("NewCustomer must reject HTML in full name")
	}
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestVehicleRejectsHTML(t *testing.T) {
	_, err := NewVehicle("v-1", "c-1", "KXR482", "VIN12345678901234", "<script>bad</script>", "Model", 2020, time.Now())
	if err == nil {
		t.Fatal("NewVehicle must reject HTML in brand")
	}
}

func TestServiceOrderRejectsHTML(t *testing.T) {
	_, err := NewServiceOrder("so-1", "OS-0001", "v-1", "Falla con <b>html</b>", time.Now())
	if err == nil {
		t.Fatal("NewServiceOrder must reject HTML in reported failure")
	}
}
