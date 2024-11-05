import os


def try_stuff(dir: str):
    failed = []
    files = os.listdir(dir)
    for file in files:
        if file.endswith(".v"):
            fname = os.path.join(dir, file)
            r = os.system(f"cat {fname} | ./myls")
            if r != 0:
                failed.append(fname)
    return failed


r = try_stuff("/home/eliot/Documents/GitHub/cs147/CS147-Project03") + try_stuff(
    "/home/eliot/Documents/GitHub/cs147/CS147-Project03/TESTBENCH"
)

print(r)
