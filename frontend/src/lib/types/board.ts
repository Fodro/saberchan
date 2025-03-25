import type { Thread } from "./thread";

export type Board  = {
	id: string;
	alias: string;
	name: string;
	description: string;
	locked: boolean;
	threads: Thread[];
}