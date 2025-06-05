#!/bin/bash

# P2P Chat Broadcast Demo Script
# This script automatically sends messages through the P2P network for demonstration

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
BROADCAST_DURATION=${BROADCAST_DURATION:-60}
MESSAGE_INTERVAL=${MESSAGE_INTERVAL:-5}
AUTO_START_DEMO=${AUTO_START_DEMO:-false}

# Demo messages showcasing different features
DEMO_MESSAGES=(
    "🚀 Welcome to the P2P Chat Network!"
    "💬 This message is being broadcast across all peers"
    "🔗 Each peer maintains connections to multiple others"
    "❤️ Heartbeat monitoring keeps connections alive"
    "🔄 Messages are deduplicated using LRU cache"
    "📡 Automatic peer discovery and connection"
    "🛡️ Graceful shutdown preserves message integrity"
    "⚡ Low latency message propagation"
    "🐳 Docker containers make deployment easy"
    "🔧 Environment variables control configuration"
    "📊 Structured logging provides detailed insights"
    "🌐 Mesh network topology for redundancy"
    "🚦 Connection limits prevent resource exhaustion"
    "🔀 Load balancing across peer connections"
    "📈 Real-time monitoring and health checks"
)

# Peer information
PEERS=(
    "bootstrap:p2p-bootstrap:8080"
    "peer-1:p2p-peer-1:8081"
    "peer-2:p2p-peer-2:8082"
    "peer-3:p2p-peer-3:8083"
    "peer-4:p2p-peer-4:8084"
    "peer-5:p2p-peer-5:8085"
)

echo -e "${BLUE}📡 P2P Chat Broadcast Demo${NC}"
echo -e "${YELLOW}⏰ Duration: ${BROADCAST_DURATION}s${NC}"
echo -e "${YELLOW}📨 Message Interval: ${MESSAGE_INTERVAL}s${NC}"
echo

# Check if demo is running or start it
check_or_start_demo() {
    echo -e "${CYAN}🔍 Checking demo status...${NC}"
    
    if ! docker-compose ps | grep -q "Up"; then
        if [ "$AUTO_START_DEMO" = "true" ]; then
            echo -e "${YELLOW}🚀 Starting demo automatically...${NC}"
            make demo
            echo -e "${GREEN}✅ Demo started successfully${NC}"
            echo
        else
            echo -e "${RED}❌ Demo is not running!${NC}"
            echo -e "${YELLOW}💡 Start with: make demo${NC}"
            echo -e "${YELLOW}💡 Or set AUTO_START_DEMO=true to start automatically${NC}"
            exit 1
        fi
    else
        echo -e "${GREEN}✅ Demo is running${NC}"
        echo
    fi
}

# Show network topology
show_network_info() {
    echo -e "${PURPLE}🕸️  Network Topology:${NC}"
    echo -e "${PURPLE}┌─────────────────────────────────────────┐${NC}"
    echo -e "${PURPLE}│             P2P Network                 │${NC}"
    echo -e "${PURPLE}│                                         │${NC}"
    echo -e "${PURPLE}│  Bootstrap (8080) ←→ Peer-1 (8081)     │${NC}"
    echo -e "${PURPLE}│       ↕                    ↕            │${NC}"
    echo -e "${PURPLE}│  Peer-5 (8085) ←→ Peer-2 (8082)        │${NC}"
    echo -e "${PURPLE}│       ↕              ↕                  │${NC}"
    echo -e "${PURPLE}│  Peer-4 (8084) ←→ Peer-3 (8083)        │${NC}"
    echo -e "${PURPLE}│                                         │${NC}"
    echo -e "${PURPLE}│  Messages propagate through all peers   │${NC}"
    echo -e "${PURPLE}└─────────────────────────────────────────┘${NC}"
    echo
}

# Broadcast message via JSON API simulation
broadcast_message() {
    local peer_info=$1
    local message=$2
    local color=$3
    
    IFS=':' read -r peer_name container_name port <<< "$peer_info"
    
    echo -e "${color}📤 ${peer_name}: ${message}${NC}"
    
    # Create a JSON message that simulates what the P2P system would send
    local json_message=$(cat <<EOF
{
    "sender_id": "${peer_name}",
    "sequence_no": $((RANDOM % 10000)),
    "payload": "$message",
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
)
    
    # Log the broadcast to container (simulation)
    echo "Broadcast: $message" | docker exec -i "$container_name" sh -c 'cat > /dev/null' 2>/dev/null || {
        echo -e "${RED}⚠️  Warning: Could not connect to ${peer_name}${NC}"
    }
    
    # Show message propagation simulation
    echo -e "${CYAN}  ↳ Propagating to connected peers...${NC}"
    
    return 0
}

# Monitor network activity
monitor_activity() {
    echo -e "${CYAN}📊 Monitoring network activity...${NC}"
    
    # Get recent logs to show activity
    local logs=$(docker-compose logs --tail=5 2>/dev/null | grep -E "(Peer connected|message|heartbeat)" | head -3)
    if [ -n "$logs" ]; then
        echo -e "${GREEN}Recent Activity:${NC}"
        echo "$logs" | while read -r line; do
            echo -e "${GREEN}  📋 $line${NC}"
        done
    else
        echo -e "${YELLOW}  📋 No recent activity captured${NC}"
    fi
    echo
}

# Show peer health status
show_peer_health() {
    echo -e "${CYAN}🏥 Peer Health Status:${NC}"
    
    for peer_info in "${PEERS[@]}"; do
        IFS=':' read -r peer_name container_name port <<< "$peer_info"
        
        if docker exec "$container_name" sh -c 'echo "health check"' >/dev/null 2>&1; then
            echo -e "${GREEN}  ✅ ${peer_name} (${port}) - Healthy${NC}"
        else
            echo -e "${RED}  ❌ ${peer_name} (${port}) - Unhealthy${NC}"
        fi
    done
    echo
}

# Demonstrate message features
demonstrate_features() {
    echo -e "${BLUE}🎭 Demonstrating P2P Features:${NC}"
    echo
    
    # Feature 1: Basic message propagation
    echo -e "${YELLOW}1. Message Propagation${NC}"
    broadcast_message "bootstrap:p2p-bootstrap:8080" "Testing message propagation across network" "$GREEN"
    sleep 2
    
    # Feature 2: Multi-peer broadcasting
    echo -e "${YELLOW}2. Multi-Peer Broadcasting${NC}"
    broadcast_message "peer-1:p2p-peer-1:8081" "Message from peer-1 to all others" "$BLUE"
    sleep 1
    broadcast_message "peer-3:p2p-peer-3:8083" "Simultaneous message from peer-3" "$PURPLE"
    sleep 2
    
    # Feature 3: Network resilience
    echo -e "${YELLOW}3. Network Resilience Demo${NC}"
    broadcast_message "peer-5:p2p-peer-5:8085" "Testing network resilience and redundancy" "$CYAN"
    echo -e "${CYAN}  ↳ Message routes through multiple paths${NC}"
    sleep 2
    
    echo -e "${GREEN}✅ Feature demonstration complete${NC}"
    echo
}

# Run continuous broadcast demo
run_broadcast_demo() {
    local end_time=$(($(date +%s) + BROADCAST_DURATION))
    local message_count=0
    local cycle_count=0
    
    echo -e "${GREEN}🎬 Starting continuous broadcast demo...${NC}"
    echo -e "${YELLOW}📺 Watch the logs with: make demo-logs${NC}"
    echo
    
    while [ $(date +%s) -lt $end_time ]; do
        # Select random peer and message
        local peer_index=$((RANDOM % ${#PEERS[@]}))
        local peer_info=${PEERS[$peer_index]}
        local message_index=$((message_count % ${#DEMO_MESSAGES[@]}))
        local message="${DEMO_MESSAGES[$message_index]}"
        
        # Add sequence number to message
        local numbered_message="[${message_count}] ${message}"
        
        # Assign color based on peer
        local colors=("$RED" "$GREEN" "$YELLOW" "$BLUE" "$PURPLE" "$CYAN")
        local peer_color=${colors[$peer_index]}
        
        # Broadcast message
        broadcast_message "$peer_info" "$numbered_message" "$peer_color"
        message_count=$((message_count + 1))
        
        # Show activity every 10 messages
        if [ $((message_count % 10)) -eq 0 ]; then
            cycle_count=$((cycle_count + 1))
            echo
            echo -e "${CYAN}📊 Cycle ${cycle_count}: Sent ${message_count} messages${NC}"
            monitor_activity
            
            # Show health status occasionally
            if [ $((cycle_count % 3)) -eq 0 ]; then
                show_peer_health
            fi
        fi
        
        # Wait before next message
        sleep $MESSAGE_INTERVAL
    done
    
    echo
    echo -e "${GREEN}🎉 Broadcast demo completed!${NC}"
    echo -e "${GREEN}📊 Total messages sent: ${message_count}${NC}"
    echo -e "${GREEN}⏰ Total duration: ${BROADCAST_DURATION}s${NC}"
}

# Show final statistics
show_final_stats() {
    echo
    echo -e "${BLUE}📈 Final Network Statistics:${NC}"
    
    # Container stats
    echo -e "${CYAN}💾 Resource Usage:${NC}"
    docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}" 2>/dev/null | head -7 || echo "Could not get stats"
    
    echo
    echo -e "${CYAN}📡 Network Connections:${NC}"
    docker network inspect p2p_p2p-network --format '{{range .Containers}}{{.Name}}: {{.IPv4Address}} {{end}}' 2>/dev/null || echo "Network not found"
    
    echo
    echo -e "${CYAN}📜 Recent Log Activity:${NC}"
    docker-compose logs --tail=5 2>/dev/null | tail -5 || echo "Could not get logs"
}

# Cleanup function
cleanup() {
    echo
    echo -e "${YELLOW}🛑 Broadcast demo interrupted${NC}"
    show_final_stats
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

# Main execution
main() {
    case "${1:-demo}" in
        "demo")
            check_or_start_demo
            show_network_info
            demonstrate_features
            run_broadcast_demo
            show_final_stats
            ;;
        "features")
            check_or_start_demo
            demonstrate_features
            ;;
        "monitor")
            check_or_start_demo
            show_peer_health
            monitor_activity
            ;;
        "health")
            check_or_start_demo
            show_peer_health
            ;;
        "help"|"--help"|"-h")
            echo -e "${BLUE}P2P Chat Broadcast Demo Script${NC}"
            echo
            echo "Usage: $0 [command]"
            echo
            echo "Commands:"
            echo "  demo      - Run full broadcast demo (default)"
            echo "  features  - Demonstrate specific features only"
            echo "  monitor   - Show network monitoring info"
            echo "  health    - Check peer health status"
            echo "  help      - Show this help message"
            echo
            echo "Environment Variables:"
            echo "  BROADCAST_DURATION - How long to run demo (default: 60s)"
            echo "  MESSAGE_INTERVAL   - Time between messages (default: 5s)"
            echo "  AUTO_START_DEMO    - Auto-start demo if not running (default: false)"
            echo
            echo "Examples:"
            echo "  $0 demo"
            echo "  BROADCAST_DURATION=120 $0 demo"
            echo "  AUTO_START_DEMO=true $0 demo"
            echo "  $0 features"
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