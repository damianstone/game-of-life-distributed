import matplotlib.pyplot as plt
import re

#  without SDL
data = """
BenchmarkGOL/512x512x10000-1-8         	       1	223952700638 ns/op	68247624 B/op	  181982 allocs/op
BenchmarkGOL/512x512x10000-2-8         	       1	133581973270 ns/op	41257976 B/op	  111598 allocs/op
BenchmarkGOL/512x512x10000-3-8         	       1	105894012434 ns/op	33222352 B/op	   90575 allocs/op
BenchmarkGOL/512x512x10000-4-8         	       1	90681457518 ns/op	34419520 B/op	   93026 allocs/op
"""

# signal SDL
data = """
BenchmarkGOL/512x512x1000-1-8         	       1	62272204403 ns/op	22102448 B/op	   56665 allocs/op
BenchmarkGOL/512x512x1000-2-8         	       1	37184262015 ns/op	13340016 B/op	   35788 allocs/op
BenchmarkGOL/512x512x1000-3-8         	       1	30277619373 ns/op	11440760 B/op	   31332 allocs/op
BenchmarkGOL/512x512x1000-4-8         	       1	27861971582 ns/op	11061440 B/op	   29605 allocs/op
"""

# Regular expression pattern to extract relevant data
pattern = r"BenchmarkGOL/(\d+)x(\d+)x(\d+)-(\d+)-\d+\s+\d+\s+(\d+) ns/op\s+(\d+) B/op\s+(\d+) allocs/op"
matches = re.findall(pattern, data)

# Extracted data
threads = [int(match[3]) for match in matches]
time_ns = [int(match[4]) for match in matches]

# Convert time to seconds
time_sec = [t / 1e9 for t in time_ns]

# Create the plot
plt.figure(figsize=(10, 6))
plt.bar(threads, time_sec, width=0.4)
# plt.plot(threads, time_sec, marker='o', linestyle='-')
plt.xlabel('Number of Nodes')
plt.ylabel('Time (seconds)')
plt.title('Distributed Implementation (AWS t2.xlarge)')
plt.xticks(threads)
plt.grid(axis='y', linestyle='--', alpha=0.7)

# Display the plot
plt.show()