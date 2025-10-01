export interface InputBuffer {
  left: boolean;
  right: boolean;
  rotate: boolean; //rotate clockwise with arrow up
  rotateCounterClockwise: boolean;
  down: boolean;
  space: boolean;
  hold: boolean;

  //botActions: BotAction[];
}

export type WsMessage = {
  type: string;
  //to?: string; //might use in future
  payload: {
    listBlock?: number[]; //opponent blocks
    startAt?: number; //millisecond from UNIX Epoch
    key?: string;
    timestamp?: number;
  };
};
