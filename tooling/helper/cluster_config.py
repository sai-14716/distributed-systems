#!/usr/bin/env python3
"""
Distributed Cluster Config Helper

This module provides config loading and endpoint resolution for Python scripts.

Usage:
    from tooling.helper.cluster_config import Config
    cfg = Config.load()
    
    # Get a node's Raft endpoint
    url = cfg.get_node_url("node1", "/state")  # -> http://IP:port/state
    
    # Get a backend endpoint  
    url = cfg.get_backend_url("backend-1")     # -> http://IP:port/
    
    # Get all nodes
    for node_name in cfg.get_node_names():
        ...
    
    # Get discovery service
    url = cfg.get_discovery_url("http")        # -> http://IP:port
    udp_addr = cfg.get_discovery_url("udp")    # -> IP:port
"""

import os
import sys
from pathlib import Path
from typing import Dict, List, Optional, Any
import json

try:
    import yaml
except ImportError:
    yaml = None


class Config:
    """Distributed cluster topology configuration"""
    
    _instance = None
    
    def __init__(self, config_data: Dict[str, Any]):
        self.config = config_data
        self.laptops = config_data.get("laptops", {})
        self.nodes = config_data.get("nodes", {})
        self.backends = config_data.get("backends", {})
        self.discovery = config_data.get("discovery", {})
    
    @staticmethod
    def _find_config_file() -> Path:
        """Find cluster_config.yaml in the repo root"""
        # Try from current working directory
        pwd = Path.cwd()
        for _ in range(5):  # Look up to 5 levels
            candidate = pwd / "cluster_config.yaml"
            if candidate.exists():
                return candidate
            pwd = pwd.parent
        
        # Try relative to this file
        script_dir = Path(__file__).parent
        candidate = script_dir.parent.parent / "cluster_config.yaml"
        if candidate.exists():
            return candidate
        
        raise FileNotFoundError("cluster_config.yaml not found")
    
    @staticmethod
    def load(config_file: Optional[str] = None) -> "Config":
        """Load cluster configuration from YAML file"""
        if config_file is None:
            config_file = Config._find_config_file()
        else:
            config_file = Path(config_file)
        
        if not config_file.exists():
            raise FileNotFoundError(f"Config file not found: {config_file}")
        
        if yaml is None:
            raise ImportError(
                "pyyaml is required. Install with: pip install pyyaml"
            )
        
        with open(config_file) as f:
            config_data = yaml.safe_load(f)
        
        return Config(config_data)
    
    def get_node_url(self, node_name: str, endpoint: str = "") -> str:
        """Get the HTTP endpoint for a node's Raft service
        
        Args:
            node_name: Name of the node (e.g., "node1")
            endpoint: Optional path to append (e.g., "/state")
        
        Returns:
            Full HTTP URL (e.g., "http://192.168.1.100:19091/state")
        """
        if node_name not in self.nodes:
            raise ValueError(f"Unknown node: {node_name}")
        
        node_info = self.nodes[node_name]
        laptop = node_info.get("laptop")
        port = node_info.get("raft_port")
        
        if not laptop or not port:
            raise ValueError(f"Invalid node info for {node_name}")
        
        ip = self.laptops.get(laptop)
        if not ip:
            raise ValueError(f"Unknown laptop: {laptop}")
        
        return f"http://{ip}:{port}{endpoint}"
    
    def get_lb_url(self, node_name: str, endpoint: str = "") -> str:
        """Get the HTTP endpoint for a load balancer instance
        
        Args:
            node_name: Name of the node running the LB (e.g., "node1")
            endpoint: Optional path to append
        
        Returns:
            Full HTTP URL
        """
        if node_name not in self.nodes:
            raise ValueError(f"Unknown node: {node_name}")
        
        node_info = self.nodes[node_name]
        laptop = node_info.get("laptop")
        port = node_info.get("lb_port")
        
        if not laptop or not port:
            raise ValueError(f"Invalid LB port for {node_name}")
        
        ip = self.laptops.get(laptop)
        if not ip:
            raise ValueError(f"Unknown laptop: {laptop}")
        
        return f"http://{ip}:{port}{endpoint}"
    
    def get_backend_url(self, backend_name: str, endpoint: str = "") -> str:
        """Get the HTTP endpoint for a backend service
        
        Args:
            backend_name: Name of the backend (e.g., "backend-1")
            endpoint: Optional path to append (default: "")
        
        Returns:
            Full HTTP URL
        """
        if backend_name not in self.backends:
            raise ValueError(f"Unknown backend: {backend_name}")
        
        backend_info = self.backends[backend_name]
        laptop = backend_info.get("laptop")
        port = backend_info.get("port")
        
        if not laptop or not port:
            raise ValueError(f"Invalid backend info for {backend_name}")
        
        ip = self.laptops.get(laptop)
        if not ip:
            raise ValueError(f"Unknown laptop: {laptop}")
        
        return f"http://{ip}:{port}{endpoint}"
    
    def get_discovery_url(self, protocol: str = "http") -> str:
        """Get the discovery service endpoint
        
        Args:
            protocol: "http" or "udp"
        
        Returns:
            URL for HTTP, or "ip:port" for UDP
        """
        laptop = self.discovery.get("laptop")
        if not laptop:
            raise ValueError("Discovery service not configured")
        
        ip = self.laptops.get(laptop)
        if not ip:
            raise ValueError(f"Unknown laptop: {laptop}")
        
        if protocol.lower() in ("http", "https"):
            port = self.discovery.get("http_port")
            return f"http://{ip}:{port}"
        elif protocol.lower() == "udp":
            port = self.discovery.get("udp_port")
            return f"{ip}:{port}"
        else:
            raise ValueError(f"Unknown protocol: {protocol}")
    
    def get_node_names(self) -> List[str]:
        """Get all configured node names in sorted order"""
        return sorted(self.nodes.keys())
    
    def get_backend_names(self) -> List[str]:
        """Get all configured backend names in sorted order"""
        return sorted(self.backends.keys())
    
    def get_laptop_ip(self, laptop: str) -> str:
        """Get IP address for a laptop"""
        ip = self.laptops.get(laptop)
        if not ip:
            raise ValueError(f"Unknown laptop: {laptop}")
        return ip
    
    def __repr__(self) -> str:
        return f"<Config nodes={len(self.nodes)} backends={len(self.backends)}>"


def fetch_json(url: str, timeout: float = 0.6) -> Optional[Dict]:
    """Fetch and parse JSON from a URL
    
    Args:
        url: URL to fetch
        timeout: Timeout in seconds
    
    Returns:
        Parsed JSON dict, or None on error
    """
    try:
        import urllib.request
        import json as json_lib
        
        req = urllib.request.Request(url)
        with urllib.request.urlopen(req, timeout=timeout) as response:
            return json_lib.loads(response.read().decode('utf-8'))
    except Exception:
        return None


def fetch_text(url: str, timeout: float = 0.6) -> Optional[str]:
    """Fetch text from a URL
    
    Args:
        url: URL to fetch
        timeout: Timeout in seconds
    
    Returns:
        Response text, or None on error
    """
    try:
        import urllib.request
        
        req = urllib.request.Request(url)
        with urllib.request.urlopen(req, timeout=timeout) as response:
            return response.read().decode('utf-8')
    except Exception:
        return None