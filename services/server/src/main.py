import os
import signal
import sys
from typing import Any

import logger
import server
from lottery import Lottery
from server.lottery_monitor import LotteryMonitor

SERVER_HOST = os.getenv("SERVER_HOST")
SERVER_PORT = int(os.getenv("SERVER_PORT"))
STORAGE_PATH = os.getenv("STORAGE_PATH", "./data/bets.csv")
AGENCY_QUORUM_MIN = int(os.getenv("AGENCY_QUORUM_MIN"))


def setup_lottery() -> LotteryMonitor:
    os.makedirs(os.path.dirname(STORAGE_PATH), exist_ok=True)

    if AGENCY_QUORUM_MIN is None:
        logger.error(
            "load-config",
            logger.LogResult.fail,
            "err",
            "AGENCY_QUORUM_MIN environment variable is required",
        )
        raise ValueError("AGENCY_QUORUM_MIN environment variable is required")

    logger.info("init-storage", logger.LogResult.success, "path", STORAGE_PATH)

    return LotteryMonitor(Lottery(STORAGE_PATH), AGENCY_QUORUM_MIN)



def main():
    logger.init()

    try:
        lottery_monitor = setup_lottery()
    except Exception as e:
        logger.error("setup", logger.LogResult.fail, "err", str(e))
        return 1

    s = server.Server(SERVER_HOST, SERVER_PORT, lottery_monitor)

    def handle_shutdown_signal(signum: Any, frame: Any) -> None:
        logger.info("shutdown-signal", logger.LogResult.in_progress, "signal", signum)
        s.shutdown()

    signal.signal(signal.SIGTERM, handle_shutdown_signal)
    signal.signal(signal.SIGINT, handle_shutdown_signal)

    s.start()
    s.join()

    return 0

if __name__ == "__main__":
    sys.exit(main())
