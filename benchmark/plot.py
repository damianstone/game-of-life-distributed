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
BenchmarkGOL/512x512x1000-1-8         	       1	28586170509 ns/op	1240984152 B/op	 6677222 allocs/op
BenchmarkGOL/512x512x1000-2-8         	       1	32432511364 ns/op	1240731360 B/op	 6679262 allocs/op
BenchmarkGOL/512x512x1000-3-8         	       1	32878949288 ns/op	1240703280 B/op	 6679163 allocs/op
BenchmarkGOL/512x512x1000-4-8         	       1	35298024910 ns/op	1241282536 B/op	 6680720 allocs/op
"""

# no signal SDL
data = """
BenchmarkGOL/512x512x1000-1-8         	       1	39972179001 ns/op	1244428744 B/op	 6686566 allocs/op
BenchmarkGOL/512x512x1000-2-8         	       1	33282712288 ns/op	1241287032 B/op	 6680800 allocs/op
BenchmarkGOL/512x512x1000-3-8         	       1	30220771037 ns/op	1222031320 B/op	 5661488 allocs/op
BenchmarkGOL/512x512x1000-4-8         	       1	28543222902 ns/op	1240144736 B/op	 6677703 allocs/op
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