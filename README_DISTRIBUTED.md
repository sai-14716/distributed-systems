# 🎯 Distributed Systems Multi-Machine Cluster - Start Here

Welcome! This repository has been converted from a single-machine Docker Compose setup to a **distributed multi-machine Raft consensus cluster** with load balancing.

## 📋 Quick Navigation

### For First-Time Users

👉 **Start here**: [DISTRIBUTED_SETUP.md](DISTRIBUTED_SETUP.md)

- Overview of the new architecture
- Step-by-step setup instructions
- How to configure your machine IPs
- Running your first demo

### For Quick Reference

👉 **Commands cheat sheet**: [QUICK_REFERENCE.md](QUICK_REFERENCE.md)

- All important commands
- Common troubleshooting
- Helper function examples

### For Technical Details

👉 **What changed**: [CONVERSION_SUMMARY.md](CONVERSION_SUMMARY.md)

- Complete list of file modifications
- Explanation of key design changes
- Breaking changes and migration guide

## 🚀 30-Second Setup

```bash
# 1. Edit cluster_config.yaml with your machine IPs
edit cluster_config.yaml
# Change LAPTOP_A_IP to your Laptop A's IP address

# 2. Start the cluster
bash tooling/demo/up.sh

# 3. View cluster state
bash tooling/demo/state.sh
```

## 🏗️ Architecture

### Before (Single Machine)

```
Laptop A
├── All nodes (1-5)
├── All backends (1-10)
└── All communication via localhost
```

### After (Multi-Machine)

```
Laptop A                    Laptop B                Laptop C
├── node1, node2           ├── node3, node4         ├── node5
├── backend-1/2/3          ├── backend-4/5/6/7      ├── backend-8/9/10
├── discovery service      └── (optional)           └── (optional)
└── LBs on ports 8001-8005
```

**All communication uses real IP addresses and exposed ports**

## 📂 Key Files

| File                               | Purpose                                        |
| ---------------------------------- | ---------------------------------------------- |
| `cluster_config.yaml`              | Cluster topology (IPs, ports, service mapping) |
| `docker-compose.yaml`              | Laptop A services only                         |
| `docker-compose.laptopb.yaml`      | Laptop B template (optional)                   |
| `docker-compose.laptopc.yaml`      | Laptop C template (optional)                   |
| `tooling/helper/cluster_config.sh` | Bash helper for config loading                 |
| `tooling/helper/cluster_config.py` | Python helper for config loading               |
| `tooling/demo/*.sh`                | Refactored demo scripts                        |

## ✨ What's New

✅ **Configuration-driven**: All topology from `cluster_config.yaml`  
✅ **Topology-aware scripts**: Scripts resolve endpoints dynamically  
✅ **Real network communication**: Uses machine IPs, not container names  
✅ **Scalable**: Easy to add more laptops/services  
✅ **Helper modules**: Reusable config loader for bash & Python

## ❓ Common Questions

**Q: Can I run this on a single machine?**  
A: Yes! Just set all laptop IPs to `localhost` in cluster_config.yaml, or use a single machine's IP.

**Q: Do I need all 3 laptops?**  
A: No. Laptop A is fully functional standalone. Add B and C as needed.

**Q: Which ports are used?**  
A: See [DISTRIBUTED_SETUP.md - Port Reference](DISTRIBUTED_SETUP.md#port-reference)

**Q: How do I update the IPs?**  
A: Edit `cluster_config.yaml` and update `LAPTOP_A`, `LAPTOP_B`, `LAPTOP_C` sections.

**Q: Do scripts still work?**  
A: Yes! But their parameters changed. See [QUICK_REFERENCE.md](QUICK_REFERENCE.md)

## 🔨 Technology Stack

- **Consensus**: Raft protocol (Go implementation)
- **Load Balancing**: Multiple algorithms (Round-Robin, Maglev, Least-Requests, etc.)
- **Service Discovery**: DNS-based with health probes
- **Container Runtime**: Docker Compose
- **Scripting**: Bash + Python 3
- **Configuration**: YAML

## 📚 Documentation

- **[DISTRIBUTED_SETUP.md](DISTRIBUTED_SETUP.md)** - Complete setup guide
- **[QUICK_REFERENCE.md](QUICK_REFERENCE.md)** - Command reference & troubleshooting
- **[CONVERSION_SUMMARY.md](CONVERSION_SUMMARY.md)** - Technical details of changes
- **[DEMO.md](docs/DEMO.md)** - Original demo documentation

## ⚡ Next Steps

1. **Read** [DISTRIBUTED_SETUP.md](DISTRIBUTED_SETUP.md)
2. **Configure** `cluster_config.yaml` with your IPs
3. **Start** with `bash tooling/demo/up.sh`
4. **Explore** demos: `state.sh`, `show_routing_headers.sh`, `watch_health_propagation.sh`
5. **Optional**: Set up laptops B and C with provided templates

## 🆘 Getting Help

- **Setup issues**: See [Troubleshooting](DISTRIBUTED_SETUP.md#troubleshooting)
- **Script changes**: See [Quick Reference](QUICK_REFERENCE.md)
- **Technical details**: See [Conversion Summary](CONVERSION_SUMMARY.md)
- **Log viewing**: `docker compose logs -f [service_name]`
- **Manual testing**: `curl http://[ip]:[port]/[endpoint]`

## 📝 Notes

- All IPs in `cluster_config.yaml` must match actual machine IPs
- Configuration must be consistent across all machines
- Firewall may need to allow ports 19091-19095, 8001-8090, 6699, 6701
- Services bind to `0.0.0.0` (all interfaces), not just localhost
- No changes needed to Go/Rust code - only infrastructure & scripts

---

**Ready to get started?** → [DISTRIBUTED_SETUP.md](DISTRIBUTED_SETUP.md)
