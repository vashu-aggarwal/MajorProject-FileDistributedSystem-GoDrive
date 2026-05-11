import urllib.request
import urllib.parse
import urllib.error
import base64
import os
import time
import json
import mimetypes

URL = "http://localhost:4000"

def encode_multipart_formdata(fields, files):
    boundary = '----WebKitFormBoundary7MA4YWxkTrZu0gW'
    CRLF = b'\r\n'
    L = []
    
    for key, value in fields.items():
        L.append(f'--{boundary}'.encode())
        L.append(f'Content-Disposition: form-data; name="{key}"'.encode())
        L.append(b'')
        L.append(str(value).encode())
        
    for key, (filename, value) in files.items():
        L.append(f'--{boundary}'.encode())
        L.append(f'Content-Disposition: form-data; name="{key}"; filename="{filename}"'.encode())
        mimetype = mimetypes.guess_type(filename)[0] or 'application/octet-stream'
        L.append(f'Content-Type: {mimetype}'.encode())
        L.append(b'')
        L.append(value)
        
    L.append(f'--{boundary}--'.encode())
    L.append(b'')
    body = CRLF.join(L)
    content_type = f'multipart/form-data; boundary={boundary}'
    return content_type, body

def test_pipeline():
    print("Starting API test...")
    
    # 1. Create a test file (binary/text mix)
    original_data = b"Hello, World! This is a test file to ensure Huffman and AES work flawlessly." + os.urandom(100)
    filename = f"test_file_{int(time.time())}.bin"
    
    print(f"Original Data Size: {len(original_data)} bytes")
    
    # 2. Upload the file
    print(f"\nUploading {filename}...")
    fields = {'fileName': filename}
    files = {'file': (filename, original_data)}
    
    content_type, body = encode_multipart_formdata(fields, files)
    
    req = urllib.request.Request(f"{URL}/upload", data=body)
    req.add_header('Content-Type', content_type)
    
    try:
        response = urllib.request.urlopen(req)
        print(f"Status Code: {response.getcode()}")
        print(f"Response: {response.read().decode('utf-8')}")
    except urllib.error.URLError as e:
        print(f"Upload failed! {e}")
        if hasattr(e, 'read'):
            print(e.read().decode('utf-8'))
        return

    # 3. Download the file
    print(f"\nDownloading {filename}...")
    req = urllib.request.Request(f"{URL}/download?filename={filename}")
    try:
        response = urllib.request.urlopen(req)
        print(f"Status Code: {response.getcode()}")
        resp_body = response.read().decode('utf-8')
    except urllib.error.URLError as e:
        print(f"Download failed! {e}")
        if hasattr(e, 'read'):
            print(e.read().decode('utf-8'))
        return
        
    json_resp = json.loads(resp_body)
    base64_content = json_resp.get("content", "")
    
    # 4. Decode base64
    downloaded_data = base64.b64decode(base64_content)
    print(f"Downloaded Data Size: {len(downloaded_data)} bytes")
    
    # 5. Verify
    if original_data == downloaded_data:
        print("\nSUCCESS: The downloaded data perfectly matches the original data!")
    else:
        print("\nFAILURE: The downloaded data does not match the original data!")

if __name__ == "__main__":
    test_pipeline()

