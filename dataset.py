import numpy as np

SEQ_MAX_LEN = 64 # TODO: move to config file, and actually determine value instead of randomly guessing

class Dataset():
	def __init__(self, files):
		self.sequences = []

		if not isinstance(files, list):
			files = [files]
		for f in files:
			self.parse_file(f)
		print self.sequences

	def parse_file(self, fp):
		#
		# Read a file line by line, append
		# its sequence of encoded bytes
		# to self.sequences
		#

		with open(fp, "r") as f:
			for line in f:
				# Truncate line to SEQ_MAX_LEN
				line = line[:SEQ_MAX_LEN-1]
				line += b'\0'
				encoded = np.zeros((SEQ_MAX_LEN, 256), dtype='bool')
				for i, char in enumerate(line):
					encoded[(i, ord(char))] = 1
				self.sequences.append(encoded)

if __name__ == '__main__':
	Dataset("test.txt")