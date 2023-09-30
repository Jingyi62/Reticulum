import pyshark

def process_packet(pkt):
    try:
        ip_src = pkt.ip.src
        ip_dst = pkt.ip.dst
        port_src = pkt[pkt.transport_layer].srcport
        port_dst = pkt[pkt.transport_layer].dstport
        pkt_length = int(pkt.length)
        if (ip_src == "127.0.0.1" and port_src == "9002"):
            # Upload traffic
            upload_stats.append(pkt_length)
        elif (ip_dst == "127.0.0.1" and port_dst == "9002"):
            # Download traffic
            download_stats.append(pkt_length)

    except AttributeError:
        pass

# 打开PCAP文件
cap = pyshark.FileCapture('temp.pcap')

# 上传和下载数据包大小统计
upload_stats = []
download_stats = []

# 逐个处理数据包
for pkt in cap:
    process_packet(pkt)

# 计算上传和下载数据包大小总和
total_upload = sum(upload_stats)
total_download = sum(download_stats)

# 输出统计结果
print("Total Upload Size:", total_upload)
print("Total Download Size:", total_download)
