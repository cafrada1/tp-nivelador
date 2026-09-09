import signal
import sys
from typing import Any

import logger
import server
from server import Config


def main():
    logger.init()

    try:
        config: Config = server.load_config()
        s = server.Server(config.server_host, config.server_port, config.lottery_monitor)
    except Exception as e:
        logger.error("server-init", logger.LogResult.fail, "err", str(e))
        return 1

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
