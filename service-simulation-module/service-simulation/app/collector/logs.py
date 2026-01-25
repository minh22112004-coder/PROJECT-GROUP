def collector():
    with open('/logs/inetsim.log') as log_file:
        for line in log_file:
            if "DNS" in line or "HTTP" in line:
                print("Event: ", line.strip())
                