#!/usr/bin/env python3
import os
import sys
import yaml

def get_config():
    root_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    config_path = os.path.join(root_dir, 'cluster_config.yaml')
    try:
        with open(config_path, 'r') as f:
            return yaml.safe_load(f)
    except FileNotFoundError:
        print(f"Error: {config_path} not found. Please create it.", file=sys.stderr)
        sys.exit(1)

def resolve(service_type, name, cfg=None):
    if cfg is None:
        cfg = get_config()
    
    if service_type == "discovery":
        info = cfg["discovery"]
        laptop_id = info["laptop"]
        ip = cfg["laptops"][laptop_id]
        return ip, info
        
    if service_type not in cfg:
        raise ValueError(f"Unknown service type: {service_type}")
    if name not in cfg[service_type]:
        raise ValueError(f"Unknown service name: {name} in {service_type}")
    
    info = cfg[service_type][name]
    laptop_id = info['laptop']
    if laptop_id not in cfg['laptops']:
        raise ValueError(f"Unknown laptop ID: {laptop_id}")
    
    ip = cfg['laptops'][laptop_id]
    return ip, info

def get_url(service_type, name, cfg=None):
    ip, info = resolve(service_type, name, cfg)
    if service_type == "discovery":
        return f"http://{ip}:{info['http_port']}"
    port = info.get('port')
    if not port:
        raise ValueError(f"No port found for {service_type}/{name}")
    return f"http://{ip}:{port}"

def main():
    if len(sys.argv) < 2:
        print(f"Usage: {sys.argv[0]} <command> [args...]")
        sys.exit(1)
        
    cmd = sys.argv[1]
    cfg = get_config()
    
    try:
        if cmd == "get_url":
            print(get_url(sys.argv[2], sys.argv[3], cfg))
        elif cmd == "get_all":
            service_type = sys.argv[2]
            print(" ".join(cfg.get(service_type, {}).keys()))
        elif cmd == "get_all_urls":
            service_type = sys.argv[2]
            urls = [get_url(service_type, name, cfg) for name in cfg.get(service_type, {}).keys()]
            print(" ".join(urls))
        elif cmd == "get_discovery_http":
            ip, info = resolve("discovery", "", cfg)
            print(f"http://{ip}:{info['http_port']}")
        elif cmd == "get_discovery_udp":
            ip, info = resolve("discovery", "", cfg)
            print(f"{ip}:{info['udp_port']}")
        else:
            print(f"Unknown command: {cmd}", file=sys.stderr)
            sys.exit(1)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == '__main__':
    main()
