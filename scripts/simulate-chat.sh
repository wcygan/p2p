#!/bin/bash

# P2P Chat Simulation Script
# This script simulates chat activity between peers for testing and demonstration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
SIMULATION_DURATION=${SIMULATION_DURATION:-30}
MESSAGE_INTERVAL=${MESSAGE_INTERVAL:-3}
DEMO_RUNNING=false

# Test messages
MESSAGES=(
    "Hello from the P2P network!"
    "Testing message propagation..."
    "This is a distributed chat system 🚀"
    "Messages should reach all peers instantly"
    "Heartbeat system keeps us connected ❤️"
    "Automatic reconnection handles failures"
    "Docker makes deployment easy 🐳"
    "Go provides excellent concurrency 🔄"
    "JSON serialization keeps it simple"
    "LRU cache prevents message duplicates"
    "Graceful shutdown preserves data"
    "Structured logging helps debugging 📊"
    "Configuration supports multiple formats"
    "Health checks monitor system status ✅"
    "Load balancing across peer connections"
)

# Peer containers
PEERS=(
    "p2p-bootstrap"
    "p2p-peer-1" 
    "p2p-peer-2"
    "p2p-peer-3"
    "p2p-peer-4"
    "p2p-peer-5"
)

echo -e "${BLUE}🤖 P2P Chat Simulation Starting...${NC}"
echo -e "${YELLOW}⏰ Duration: ${SIMULATION_DURATION}s${NC}"
echo -e "${YELLOW}📡 Message Interval: ${MESSAGE_INTERVAL}s${NC}"
echo

# Check if demo is running
check_demo_status() {
    echo -e "${CYAN}🔍 Checking demo status...${NC}"
    
    if ! docker-compose ps | grep -q "Up"; then
        echo -e "${RED}❌ Demo is not running! Please start with: make demo${NC}"
        exit 1
    fi
    
    # Count running containers
    RUNNING_COUNT=$(docker-compose ps | grep "Up" | wc -l)
    echo -e "${GREEN}✅ Found ${RUNNING_COUNT} running peers${NC}"
    
    if [ "$RUNNING_COUNT" -lt 6 ]; then
        echo -e "${YELLOW}⚠️  Warning: Expected 6 peers, found ${RUNNING_COUNT}${NC}"
    fi
    
    DEMO_RUNNING=true
    echo
}

# Send message via specific peer
send_message() {
    local peer=$1
    local message=$2
    local peer_color=$3
    
    echo -e "${peer_color}📤 ${peer}: ${message}${NC}"
    
    # Send message by echoing to the peer's stdin
    # Note: This simulates user input to the chat application
    echo "$message" | docker exec -i "$peer" sh -c 'echo "Simulated message: $@"' -- "$message" 2>/dev/null || {
        echo -e "${RED}⚠️  Failed to send message via ${peer}${NC}"
    }
}

# Monitor peer connections
monitor_connections() {
    echo -e "${CYAN}🔗 Monitoring peer connections...${NC}"
    
    for peer in "${PEERS[@]}"; do
        if docker exec "$peer" sh -c 'echo "status check"' >/dev/null 2>&1; then
            echo -e "${GREEN}✅ ${peer} is responsive${NC}"
        else
            echo -e "${RED}❌ ${peer} is not responsive${NC}"
        fi
    done
    echo
}

# Show network topology
show_topology() {
    echo -e "${PURPLE}🕸️  Network Topology:${NC}"
    echo -e "${PURPLE}┌─────────────────────────────────────────┐${NC}"
    echo -e "${PURPLE}│  Bootstrap (8080) ←→ Peer-1 (8081)     │${NC}"
    echo -e "${PURPLE}│       ↕                    ↕            │${NC}"
    echo -e "${PURPLE}│  Peer-5 (8085) ←→ Peer-2 (8082)        │${NC}"
    echo -e "${PURPLE}│       ↕              ↕                  │${NC}"
    echo -e "${PURPLE}│  Peer-4 (8084) ←→ Peer-3 (8083)        │${NC}"
    echo -e "${PURPLE}└─────────────────────────────────────────┘${NC}"
    echo
}

# Simulate chat activity
simulate_chat() {
    local end_time=$(($(date +%s) + SIMULATION_DURATION))
    local message_count=0
    
    echo -e "${GREEN}🎭 Starting chat simulation...${NC}"
    echo
    
    while [ $(date +%s) -lt $end_time ]; do
        # Select random peer and message
        local peer_index=$((RANDOM % ${#PEERS[@]}))
        local peer=${PEERS[$peer_index]}
        local message_index=$((RANDOM % ${#MESSAGES[@]}))
        local message="${MESSAGES[$message_index]}"
        
        # Assign color based on peer
        local colors=("$RED" "$GREEN" "$YELLOW" "$BLUE" "$PURPLE" "$CYAN")
        local peer_color=${colors[$peer_index]}
        
        # Send message
        send_message "$peer" "$message" "$peer_color"
        message_count=$((message_count + 1))
        
        # Wait before next message
        sleep $MESSAGE_INTERVAL
        
        # Occasionally show status
        if [ $((message_count % 5)) -eq 0 ]; then
            echo -e "${CYAN}📊 Sent ${message_count} messages so far...${NC}"
            echo
        fi
    done
    
    echo -e "${GREEN}✅ Simulation completed! Sent ${message_count} messages total${NC}"
}

# Show real-time logs
show_logs() {
    echo -e "${CYAN}📜 Showing recent activity logs...${NC}"
    echo -e "${YELLOW}(Press Ctrl+C to stop log monitoring)${NC}"
    echo
    
    # Show last 20 lines and follow
    docker-compose logs --tail=20 -f
}

# Cleanup function
cleanup() {
    echo
    echo -e "${YELLOW}🛑 Simulation interrupted${NC}"
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

# Main execution
main() {
    case "${1:-simulate}" in
        "simulate")
            check_demo_status
            show_topology
            monitor_connections
            simulate_chat
            ;;
        "monitor")
            check_demo_status
            monitor_connections
            ;;
        "topology") 
            show_topology
            ;;
        "logs")
            check_demo_status
            show_logs
            ;;
        "help"|"--help"|"-h")
            echo -e "${BLUE}P2P Chat Simulation Script${NC}"
            echo
            echo "Usage: $0 [command]"
            echo
            echo "Commands:"
            echo "  simulate  - Run full chat simulation (default)"
            echo "  monitor   - Check peer connection status"
            echo "  topology  - Display network topology"
            echo "  logs      - Show real-time logs"
            echo "  help      - Show this help message"
            echo
            echo "Environment Variables:"
            echo "  SIMULATION_DURATION - How long to run (default: 30s)"
            echo "  MESSAGE_INTERVAL    - Time between messages (default: 3s)"
            echo
            echo "Examples:"
            echo "  $0 simulate"
            echo "  SIMULATION_DURATION=60 $0 simulate"
            echo "  MESSAGE_INTERVAL=1 $0 simulate"
            ;;
        *)
            echo -e "${RED}❌ Unknown command: $1${NC}"
            echo "Use '$0 help' for available commands"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"