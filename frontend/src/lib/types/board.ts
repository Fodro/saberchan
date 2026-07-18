import type { Thread } from "./thread";

export type Board = {
	id: string;
	alias: string;
	name: string;
	description: string;
	locked: boolean;
	author: string;
	threads: Thread[];
	total_threads?: number;
	limit?: number;
	offset?: number;
};
