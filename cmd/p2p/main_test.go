package main

import (
	"strings"
	"testing"
)

func TestAddrList(t *testing.T) {
	var addrs addrList
	
	// Test initial state
	if addrs.String() != "" {
		t.Fatalf("expected empty string for empty list, got %s", addrs.String())
	}
	
	// Test adding addresses
	if err := addrs.Set("localhost:8080"); err != nil {
		t.Fatalf("error setting address: %v", err)
	}
	
	if err := addrs.Set("localhost:8081"); err != nil {
		t.Fatalf("error setting second address: %v", err)
	}
	
	// Test string representation
	expected := "localhost:8080,localhost:8081"
	if addrs.String() != expected {
		t.Fatalf("expected %s, got %s", expected, addrs.String())
	}
	
	// Verify the slice contains the addresses
	if len(addrs) != 2 {
		t.Fatalf("expected 2 addresses, got %d", len(addrs))
	}
	
	if addrs[0] != "localhost:8080" {
		t.Fatalf("expected first address localhost:8080, got %s", addrs[0])
	}
	
	if addrs[1] != "localhost:8081" {
		t.Fatalf("expected second address localhost:8081, got %s", addrs[1])
	}
}

func TestAddrListMultipleSets(t *testing.T) {
	var addrs addrList
	
	addresses := []string{
		"192.168.1.1:8080",
		"192.168.1.2:8080", 
		"192.168.1.3:8080",
	}
	
	for _, addr := range addresses {
		if err := addrs.Set(addr); err != nil {
			t.Fatalf("error setting address %s: %v", addr, err)
		}
	}
	
	result := addrs.String()
	expected := strings.Join(addresses, ",")
	
	if result != expected {
		t.Fatalf("expected %s, got %s", expected, result)
	}
}

func TestAddrListEmpty(t *testing.T) {
	var addrs addrList
	
	// Test that empty list returns empty string
	if addrs.String() != "" {
		t.Fatalf("expected empty string for empty list, got '%s'", addrs.String())
	}
}