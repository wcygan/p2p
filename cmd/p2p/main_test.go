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

func TestAddrListNilSlice(t *testing.T) {
	var addrs addrList
	
	// Test Set method on nil slice
	if err := addrs.Set("localhost:8080"); err != nil {
		t.Fatalf("error setting address on nil slice: %v", err)
	}
	
	if len(addrs) != 1 {
		t.Fatalf("expected 1 address after Set, got %d", len(addrs))
	}
	
	if addrs[0] != "localhost:8080" {
		t.Fatalf("expected address 'localhost:8080', got %s", addrs[0])
	}
}

func TestAddrListSingleAddress(t *testing.T) {
	var addrs addrList
	
	if err := addrs.Set("192.168.1.100:9090"); err != nil {
		t.Fatalf("error setting address: %v", err)
	}
	
	result := addrs.String()
	expected := "192.168.1.100:9090"
	
	if result != expected {
		t.Fatalf("expected %s, got %s", expected, result)
	}
}

func TestAddrListWithEmptyStrings(t *testing.T) {
	var addrs addrList
	
	// Add empty string
	if err := addrs.Set(""); err != nil {
		t.Fatalf("error setting empty address: %v", err)
	}
	
	// Add valid address
	if err := addrs.Set("localhost:8080"); err != nil {
		t.Fatalf("error setting valid address: %v", err)
	}
	
	result := addrs.String()
	expected := ",localhost:8080"
	
	if result != expected {
		t.Fatalf("expected %s, got %s", expected, result)
	}
}

func TestAddrListWithSpecialCharacters(t *testing.T) {
	var addrs addrList
	
	specialAddrs := []string{
		"[::1]:8080",           // IPv6 address
		"localhost:8080",       // regular address  
		"192.168.1.1:8080",    // IPv4 address
		"example.com:443",      // domain name
	}
	
	for _, addr := range specialAddrs {
		if err := addrs.Set(addr); err != nil {
			t.Fatalf("error setting address %s: %v", addr, err)
		}
	}
	
	result := addrs.String()
	expected := strings.Join(specialAddrs, ",")
	
	if result != expected {
		t.Fatalf("expected %s, got %s", expected, result)
	}
	
	if len(addrs) != len(specialAddrs) {
		t.Fatalf("expected %d addresses, got %d", len(specialAddrs), len(addrs))
	}
}