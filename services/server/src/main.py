import os
import sys

import logger
import server

SERVER_HOST = os.getenv("SERVER_HOST")
SERVER_PORT = int(os.getenv("SERVER_PORT"))
STORAGE_PATH = os.getenv("STORAGE_PATH", "./bets.csv")


def main():
    logger.init()
    s = server.Server(SERVER_HOST, SERVER_PORT, STORAGE_PATH)
    try:
        s.run()
    except Exception as e:
        logger.error("server-run", logger.LogResult.fail, "err", e)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
