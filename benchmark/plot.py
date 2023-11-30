import matplotlib.pyplot as plt
import re

# 10000 iterations
data = """
BenchmarkGOL/512x512x10000-1-8         	       1	223952700638 ns/op	68247624 B/op	  181982 allocs/op
BenchmarkGOL/512x512x10000-2-8         	       1	133581973270 ns/op	41257976 B/op	  111598 allocs/op
BenchmarkGOL/512x512x10000-3-8         	       1	105894012434 ns/op	33222352 B/op	   90575 allocs/op
BenchmarkGOL/512x512x10000-4-8         	       1	90681457518 ns/op	34419520 B/op	   93026 allocs/op
"""

# 1000 iterations until 8 nodes
data = """
BenchmarkGOL/512x512x1000-1-8         	       1	25399722852 ns/op	11271992 B/op	   27258 allocs/op
BenchmarkGOL/512x512x1000-2-8         	       1	14366502938 ns/op	 7073008 B/op	   18773 allocs/op
BenchmarkGOL/512x512x1000-3-8         	       1	11525774013 ns/op	 6279112 B/op	   17300 allocs/op
BenchmarkGOL/512x512x1000-4-8         	       1	11274729648 ns/op	 5933920 B/op	   15681 allocs/op
BenchmarkGOL/512x512x1000-5-8         	       1	12410016163 ns/op	 7568304 B/op	   18109 allocs/op
BenchmarkGOL/512x512x1000-6-8         	       1	12634381848 ns/op	 6258360 B/op	   17410 allocs/op
BenchmarkGOL/512x512x1000-7-8         	       1	13116558494 ns/op	 6195048 B/op	   17325 allocs/op
BenchmarkGOL/512x512x1000-8-8         	       1	11564526608 ns/op	 7279168 B/op	   16418 allocs/op
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
plt.title('Distributed Implementation (Quad-Core Intel Core i7)')
plt.xticks(threads)
plt.grid(axis='y', linestyle='--', alpha=0.7)

# Display the plot
plt.show()