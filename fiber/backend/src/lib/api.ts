const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:3000/api';

export async function fetchApi<T>(path: string, options?: RequestInit): Promise<T> {
	const res = await fetch(`${API_BASE}${path}`, {
		...options,
		headers: {
			'Content-Type': 'application/json',
			...options?.headers
		}
	});
	if (!res.ok) {
		const err = await res.json().catch(() => ({ error: res.statusText }));
		throw new Error(err.error || res.statusText);
	}
	return res.json();
}

export const api = {
	sites: {
		list: () => fetchApi<Site[]>('/sites'),
		get: (id: number) => fetchApi<Site>(`/sites/${id}`),
		create: (data: CreateSiteReq) =>
			fetchApi<Site>('/sites', { method: 'POST', body: JSON.stringify(data) }),
		update: (id: number, data: UpdateSiteReq) =>
			fetchApi<Site>(`/sites/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
		delete: (id: number) =>
			fetch(`${API_BASE}/sites/${id}`, { method: 'DELETE' })
	},
	templates: {
		list: () => fetchApi<string[]>('/templates'),
	},
	categories: {
		tree: (parentId = 0) => fetchApi<CategoryResp[]>(`/categories/tree?parent_id=${parentId}`),
		flat: (parentId = 0) => fetchApi<CategoryResp[]>(`/categories/flat?parent_id=${parentId}`),
		get: (id: number) => fetchApi<CategoryResp>(`/categories/${id}`),
		create: (data: CreateCategoryReq) =>
			fetchApi<CategoryResp>('/categories', { method: 'POST', body: JSON.stringify(data) }),
		update: (id: number, data: UpdateCategoryReq) =>
			fetchApi<CategoryResp>(`/categories/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
		delete: (id: number) =>
			fetch(`${API_BASE}/categories/${id}`, { method: 'DELETE' })
	}
};

export interface Site {
	id: number;
	name: string;
	domain: string;
	bind?: number;
	subdomains: string[];
	template: string;
	is_default: boolean;
	status: number;
}

export interface CreateSiteReq {
	name: string;
	domain: string;
	bind?: number;
	subdomains?: string[];
	template?: string;
	is_default?: boolean;
	status?: number;
}

export interface UpdateSiteReq {
	name?: string;
	domain?: string;
	bind?: number;
	subdomains?: string[];
	template?: string;
	is_default?: boolean;
	status?: number;
}

export interface CategoryResp {
	id: number;
	parent_id: number;
	name: string;
	slug: string;
	type: string;
	link: string;
	sort: number;
	status: number;
	children?: CategoryResp[];
}

export interface CreateCategoryReq {
	parent_id?: number;
	name: string;
	slug?: string;
	type?: string;
	link?: string;
	sort?: number;
	status?: number;
}

export interface UpdateCategoryReq {
	parent_id?: number;
	name?: string;
	slug?: string;
	type?: string;
	link?: string;
	sort?: number;
	status?: number;
}
