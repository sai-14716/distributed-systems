import urllib.request
import sys

proxy = urllib.request.ProxyHandler({"http": "http://127.0.0.1:6700"})
opener = urllib.request.build_opener(proxy)
req = urllib.request.Request("http://api.service.com:8000/chat", data=b"hello world", method="POST")
try:
    with opener.open(req) as response:
        print("HEADERS:")
        for k, v in response.headers.items():
            print(f"{k}: {v}")
        print("\nBODY:")
        print(response.read().decode())
except Exception as e:
    print(f"Error: {e}")
