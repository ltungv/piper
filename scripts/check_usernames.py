import json

if __name__ == "__main__":
    names = []
    passwords = []

    with open('../.creds.json', 'r') as f:
        users_list = json.load(f)

    with open('../configs/users.txt', 'r') as f:
        lines = f.readlines()
        for line in lines:
            names.append(line.strip())

    with open('../configs/passwords.txt', 'r') as f:
        lines = f.readlines()
        for line in lines:
            passwords.append(line.strip())

    keys = users_list.keys()
    for i in range(len(names)):
        if names[i] not in keys:
            if users_list[names[i]]["password"] != passwords[i]:
                print("error at: %s" % (name))
                break

    print("finished checking")
