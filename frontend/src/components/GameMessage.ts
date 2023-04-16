import React from 'react';

enum MsgType {
	State = 'state',
	Error = 'error',
	Input = 'input',
	Action = 'action',
}

interface GameMessage {
	type: MsgType;
	data: string;
}

export type { GameMessage };
export { MsgType };
