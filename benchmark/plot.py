import matplotlib.pyplot as plt
import re



# with SDL
data = """
BenchmarkGOL/512x512x1000-1-8         	       1	33888639819 ns/op	1242717048 B/op	 6681899 allocs/op
BenchmarkGOL/512x512x1000-2-8         	       1	25123140827 ns/op	1239012088 B/op	 6674620 allocs/op
BenchmarkGOL/512x512x1000-3-8         	       1	23532063906 ns/op	1220313064 B/op	 5656791 allocs/op
BenchmarkGOL/512x512x1000-4-8         	       1	22929790875 ns/op	1239001872 B/op	 6674591 allocs/op
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