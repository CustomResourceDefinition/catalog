import sys

def split_file(input_file, output_prefix):
    with open(input_file, 'r') as f:
        file_count = 1
        output_file = open(f"{output_prefix}_{file_count}.yaml", 'w')
        for line in f:
            if line.strip() == '---':
                if output_file:
                    output_file.close()
                file_count += 1
                output_file = open(f"{output_prefix}_{file_count}.yaml", 'w')
            elif line.strip().startswith("#"):
                continue
            else:
                if output_file:
                    output_file.write(line)
        if output_file:
            output_file.close()

if __name__ == "__main__":
    if len(sys.argv) != 3:
        sys.exit(1)

    input_file = sys.argv[1]
    output_prefix = sys.argv[2]
    split_file(input_file, output_prefix)
