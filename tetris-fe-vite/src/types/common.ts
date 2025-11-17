export interface InputBuffer {
  left: boolean;
  right: boolean;
  rotate: boolean; //rotate clockwise with arrow up
  rrotate: boolean; //rotate Counter Clockwise
  down: boolean;
  space: boolean;
  hold: boolean;
  downOff?: boolean;
  //botActions: BotAction[];
}

export type WsMessage = {
  type: string;
  playerid?: string; //might use in future
  payload?: {
    startAt?: number; //millisecond from UNIX Epoch aka timestamp
    key?: string;
    board?: number[][];
    latestFrame?: number;
    listBlock?: number[];
    block?: number[][]; // active block shape
    cRow?: number;
    cCol?: number;
    inputs?: Array<{ keys: string[]; frame: number }>; // batched input events
  };
  timestamp?: number;
  error?: string;
};
