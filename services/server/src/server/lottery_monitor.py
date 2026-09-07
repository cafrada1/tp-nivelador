import threading

from lottery import Lottery, Bet


class LotteryMonitor:
    def __init__(self, lottery: Lottery, quorum: int):
        self._lottery: Lottery = lottery
        self._quorum: int = quorum
        self._cond: threading.Condition = threading.Condition()
        self._finished_agencies: int = 0
        self._winners: list[Bet] | None = None

    def store_bets(self, bets: list[Bet]) -> None:
        with self._cond:
            self._lottery.store_bets(bets)

    def winners(self, agency_id: int) -> list[Bet]:
        with self._cond:
            self._finished_agencies += 1
            if self._finished_agencies >= self._quorum and self._winners is None:
                self._winners = self._draw_winners()
                self._cond.notify_all()
            while self._winners is None:
                self._cond.wait()
            return [bet for bet in self._winners if bet.agency_id == agency_id]

    def _draw_winners(self) -> list[Bet]:
        return [bet for bet in self._lottery.load_bets() if self._lottery.has_won(bet)]
