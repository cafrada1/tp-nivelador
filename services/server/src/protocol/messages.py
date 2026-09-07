from abc import ABC
from enum import Enum

from lottery import Bet


class MessageType(Enum):
    OPEN = 0x00
    DATA = 0x01
    CLOSE = 0x02
    WINNERS = 0x03
    ACK = 0x04


class Message(ABC):
    def __init__(self, message_type: MessageType):
        self.type: MessageType = message_type

    def agency_id(self) -> int:
        raise NotImplementedError("This message type does not have an agency_id")

    def bets(self) -> list[Bet]:
        raise NotImplementedError("This message type does not have bets")


class OpenMessage(Message):
    def __init__(self, agency_id: int):
        super().__init__(MessageType.OPEN)
        self._agency_id: int = agency_id

    def agency_id(self) -> int:
        return self._agency_id


class DataMessage(Message):
    def __init__(self, bets: list[Bet]):
        super().__init__(MessageType.DATA)
        self._bets: list[Bet] = bets

    def bets(self) -> list[Bet]:
        return self._bets


class CloseMessage(Message):
    def __init__(self):
        super().__init__(MessageType.CLOSE)


class AckMessage(Message):
    def __init__(self):
        super().__init__(MessageType.ACK)
