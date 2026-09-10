import threading

from lottery import Lottery, Bet


class LotteryMonitor:
    """Coordina el sorteo entre los hilos que atienden a cada agencia.

    Un hilo que pide los ganadores se registra y, al alcanzarse el quorum
    de agencias, el que llega primero sortea una única vez y despierta al
    resto. Antes del sorteo nadie recibe resultados.
    """

    def __init__(self, lottery: Lottery, quorum: int):
        self._lottery: Lottery = lottery
        self._quorum: int = quorum
        self._cond: threading.Condition = threading.Condition()
        self._agencies: set[int] = set()
        self._winners: list[Bet] | None = None
        self._aborted: bool = False

    def store_bets(self, bets: list[Bet]) -> None:
        with self._cond:
            # Si el sorteo ya se realizó, las apuestas tardías se descartan.
            if self._winners is not None:
                return
            self._lottery.store_bets(bets)

    def winners(self, agency_id: int) -> list[Bet]:
        with self._cond:
            self._agencies.add(agency_id)
            # Solo el hilo que completa el quorum ejecuta el sorteo; el resto
            # queda esperando en la condition hasta ser notificado.
            if len(self._agencies) >= self._quorum and self._winners is None:
                self._winners = self._draw_winners()
                self._cond.notify_all()
            while self._winners is None:
                if self._aborted:
                    raise RuntimeError("lottery monitor aborted while waiting for winners")
                self._cond.wait()
            return [bet for bet in self._winners if bet.agency_id == agency_id]

    def abort(self) -> None:
        # Despierta a los hilos esperando el sorteo para que fallen en vez
        # de quedarse bloqueados al cerrar el servidor.
        with self._cond:
            self._aborted = True
            self._cond.notify_all()

    def _draw_winners(self) -> list[Bet]:
        return [bet for bet in self._lottery.load_bets() if self._lottery.has_won(bet)]
