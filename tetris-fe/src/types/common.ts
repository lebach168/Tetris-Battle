export interface InputBuffer {
  left: boolean;
  right: boolean;
  up: boolean; //rotate clockwise
  rotateCounterClockwise: boolean;
  down: boolean;
  space: boolean;
  hold: boolean;

  //botActions: BotAction[];
}

export type WSMessage = {
  type: string;
  to?: string; //might use in future
  payload: {
    listBlock?: number[] | Uint8Array;//opponent blocks
    startAt?: number;//millisecond from UNIX Epoch
    key?:string;
  };
};
