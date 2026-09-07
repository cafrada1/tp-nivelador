import threading
from typing import Callable

from lottery import Lottery, Bet


class LotteryMonitor:
    def __init__(self, lottery: Lottery):
        self.lottery: Lottery = lottery
        self.mtx: threading.Lock = threading.Lock()

    def store_bets(self, bets: list[Bet]):
        with self.mtx:
            self.lottery.store_bets(bets)

    def winners(self, agency_id: int) -> list[Bet]:
        with self.mtx:
            condition: Callable[[Bet],bool] = lambda bet: agency_id == bet.agency_id and self.lottery.has_won(bet)
            return list(filter(condition, self.lottery.load_bets()))
