import os

import logger
from lottery import Lottery
from server.lottery_monitor import LotteryMonitor

SERVER_HOST_ENV = "SERVER_HOST"
SERVER_PORT_ENV = "SERVER_PORT"
STORAGE_PATH_ENV = "STORAGE_PATH"
AGENCY_QUORUM_MIN_ENV = "AGENCY_QUORUM_MIN"
STORAGE_DEFAULT = "./data/bets.csv"

class Config:
    def __init__(self, host: str, port: int, lottery_monitor: LotteryMonitor):
        self.server_host: str = host
        self.server_port: int = port
        self.lottery_monitor: LotteryMonitor = lottery_monitor

def setup_lottery(path: str, quorum: int) -> LotteryMonitor:
    os.makedirs(os.path.dirname(path), exist_ok=True)

    logger.info("init-storage", logger.LogResult.success, "path", path)

    return LotteryMonitor(Lottery(path), quorum)

def _get_env_or_raise(key: str, default: str | None = None) -> str:
    value = os.getenv(key)
    if value is None or value == "":
        if default is None:
            raise ValueError(f"{key} environment variable is not set")
        return default
    return value

def _get_env_int_or_raise(key: str) -> int:
    value = _get_env_or_raise(key)
    if not value.isnumeric():
        raise ValueError(f"{key} environment variable must be a number")
    return int(value)

def load_config() -> Config:
    quorum: int = _get_env_int_or_raise(AGENCY_QUORUM_MIN_ENV)
    if quorum <= 0:
        raise ValueError(f"{AGENCY_QUORUM_MIN_ENV} must be greater than 0")

    storage: str = _get_env_or_raise(STORAGE_PATH_ENV, default=STORAGE_DEFAULT)

    config: Config = Config(
        host=_get_env_or_raise(SERVER_HOST_ENV),
        port=_get_env_int_or_raise(SERVER_PORT_ENV),
        lottery_monitor=setup_lottery(storage, quorum),
    )

    return config