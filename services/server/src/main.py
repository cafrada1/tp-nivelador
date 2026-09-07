import os
import sys
import time

import logger
import server
from lottery import Lottery
from server.lottery_monitor import LotteryMonitor

SERVER_HOST = os.getenv("SERVER_HOST")
SERVER_PORT = int(os.getenv("SERVER_PORT"))
STORAGE_PATH = os.getenv("STORAGE_PATH", "./data/bets.csv")
AGENCY_QUORUM_MIN = int(os.getenv("AGENCY_QUORUM_MIN"))


def main():
    logger.init()

    os.makedirs(os.path.dirname(STORAGE_PATH), exist_ok=True)

    if AGENCY_QUORUM_MIN is None:
        logger.error(
            "load-config",
            logger.LogResult.fail,
            "err",
            "AGENCY_QUORUM_MIN environment variable is required",
        )
        return 1

    logger.info("init-storage", logger.LogResult.success, "path", STORAGE_PATH)

    lottery_monitor = LotteryMonitor(Lottery(STORAGE_PATH), AGENCY_QUORUM_MIN)
    s = server.Server(SERVER_HOST, SERVER_PORT, lottery_monitor)
    s.start()
    s.join()

    return 0


if __name__ == "__main__":
    sys.exit(main())
