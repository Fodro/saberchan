import type { Post } from "./post";

export type Thread = {
	id: string;
	board_id: string;
	title: string;
	original_post: Post;
	locked: boolean;
	posts: Post[];
	is_author?: boolean;
	replies_count: number;
}