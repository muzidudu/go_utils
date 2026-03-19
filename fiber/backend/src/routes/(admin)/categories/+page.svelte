<script lang="ts">
	import { onMount } from "svelte";
	import { api, type CategoryResp, type CreateCategoryReq } from "$lib/api";
	import { Button } from "$lib/components/ui/button";
	import { Card, CardHeader, CardContent } from "$lib/components/ui/card";
	import { Input } from "$lib/components/ui/input";

	let tree = $state<CategoryResp[]>([]);
	let loading = $state(true);
	let error = $state("");
	let dialogOpen = $state(false);
	let editing = $state<CategoryResp | null>(null);
	let form = $state<CreateCategoryReq>({
		parent_id: 0,
		name: "",
		slug: "",
		type: "category",
		link: "",
		sort: 0,
		status: 1,
	});
	let parentOptions = $state<{ id: number; name: string; indent: number }[]>(
		[],
	);

	function flattenTree(
		items: CategoryResp[],
		indent = 0,
	): { id: number; name: string; indent: number }[] {
		let result: { id: number; name: string; indent: number }[] = [];
		for (const item of items) {
			result.push({ id: item.id, name: item.name, indent });
			if (item.children?.length) {
				result = result.concat(flattenTree(item.children, indent + 1));
			}
		}
		return result;
	}

	async function load() {
		loading = true;
		error = "";
		try {
			tree = await api.categories.tree(0);
			parentOptions = [
				{ id: 0, name: "根级", indent: 0 },
				...flattenTree(tree),
			];
		} catch (e) {
			error = e instanceof Error ? e.message : "加载失败";
		} finally {
			loading = false;
		}
	}

	function openCreate(parentId = 0) {
		editing = null;
		form = {
			parent_id: parentId,
			name: "",
			slug: "",
			type: "category",
			link: "",
			sort: 0,
			status: 1,
		};
		dialogOpen = true;
	}

	function openEdit(cat: CategoryResp) {
		editing = cat;
		form = {
			parent_id: cat.parent_id,
			name: cat.name,
			slug: cat.slug || "",
			type: cat.type || "category",
			link: cat.link || "",
			sort: cat.sort || 0,
			status: cat.status,
		};
		dialogOpen = true;
	}

	async function save() {
		try {
			if (editing) {
				await api.categories.update(editing.id, form);
			} else {
				await api.categories.create(form);
			}
			dialogOpen = false;
			await load();
		} catch (e) {
			error = e instanceof Error ? e.message : "保存失败";
		}
	}

	async function remove(id: number) {
		if (!confirm("确定删除？")) return;
		try {
			await api.categories.delete(id);
			await load();
		} catch (e) {
			error = e instanceof Error ? e.message : "删除失败";
		}
	}

	function renderTree(items: CategoryResp[], indent = 0) {
		return items
			.map(
				(cat) => `
				<tr class="border-b border-slate-100 hover:bg-slate-50">
					<td class="px-4 py-3" style="padding-left: ${1 + indent * 1.5}rem">${"—".repeat(indent)} ${cat.name}</td>
					<td class="px-4 py-3">${cat.slug || "-"}</td>
					<td class="px-4 py-3">${cat.type || "category"}</td>
					<td class="px-4 py-3">${cat.sort}</td>
					<td class="px-4 py-3">${cat.status === 1 ? "启用" : "禁用"}</td>
					<td class="px-4 py-3 text-right">
						<button data-add="${cat.id}" class="text-primary hover:underline">添加子级</button>
						<button data-edit="${cat.id}" class="ml-2 text-primary hover:underline">编辑</button>
						<button data-remove="${cat.id}" class="ml-2 text-red-600 hover:underline">删除</button>
					</td>
				</tr>
				${cat.children?.length ? renderTree(cat.children, indent + 1) : ""}
			`,
			)
			.join("");
	}

	onMount(load);
</script>

<svelte:head>
	<title>分类管理</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold text-slate-900">分类管理</h1>
		<Button onclick={() => openCreate(0)}>新增根分类</Button>
	</div>

	{#if error}
		<div class="rounded-md bg-red-50 p-4 text-sm text-red-700">{error}</div>
	{/if}

	<Card>
		<CardHeader>
			<h2 class="text-lg font-semibold">分类树</h2>
		</CardHeader>
		<CardContent>
			{#if loading}
				<p class="py-8 text-center text-slate-500">加载中...</p>
			{:else}
				<div class="overflow-x-auto">
					<table class="w-full text-sm">
						<thead>
							<tr class="border-b border-slate-200">
								<th class="px-4 py-3 text-left font-medium"
									>名称</th
								>
								<th class="px-4 py-3 text-left font-medium"
									>Slug</th
								>
								<th class="px-4 py-3 text-left font-medium"
									>类型</th
								>
								<th class="px-4 py-3 text-left font-medium"
									>排序</th
								>
								<th class="px-4 py-3 text-left font-medium"
									>状态</th
								>
								<th class="px-4 py-3 text-right font-medium"
									>操作</th
								>
							</tr>
						</thead>
						<tbody>
							{#each tree as cat}
								<tr
									class="border-b border-slate-100 hover:bg-slate-50"
								>
									<td class="px-4 py-3">{cat.id}</td>
									<td class="px-4 py-3">{cat.name}</td>
									<td class="px-4 py-3">{cat.slug || "-"}</td>
									<td class="px-4 py-3"
										>{cat.type || "category"}</td
									>
									<td class="px-4 py-3">{cat.sort}</td>
									<td class="px-4 py-3"
										>{cat.status === 1
											? "启用"
											: "禁用"}</td
									>
									<td class="px-4 py-3 text-right">
										<Button
											variant="ghost"
											size="sm"
											onclick={() => openCreate(cat.id)}
											>添加子级</Button
										>
										<Button
											variant="ghost"
											size="sm"
											onclick={() => openEdit(cat)}
											>编辑</Button
										>
										<Button
											variant="ghost"
											size="sm"
											onclick={() => remove(cat.id)}
											>删除</Button
										>
									</td>
								</tr>
								{#if cat.children?.length}
									{#each cat.children as child}
										<tr
											class="border-b border-slate-100 bg-slate-50/50 hover:bg-slate-50"
										>
											<td class="px-4 py-3">{child.id}</td
											>
											<td class="px-4 py-3 pl-8"
												>— {child.name}</td
											>
											<td class="px-4 py-3"
												>{child.slug || "-"}</td
											>
											<td class="px-4 py-3"
												>{child.type || "category"}</td
											>
											<td class="px-4 py-3"
												>{child.sort}</td
											>
											<td class="px-4 py-3"
												>{child.status === 1
													? "启用"
													: "禁用"}</td
											>
											<td class="px-4 py-3 text-right">
												<Button
													variant="ghost"
													size="sm"
													onclick={() =>
														openCreate(child.id)}
													>添加子级</Button
												>
												<Button
													variant="ghost"
													size="sm"
													onclick={() =>
														openEdit(child)}
													>编辑</Button
												>
												<Button
													variant="ghost"
													size="sm"
													onclick={() =>
														remove(child.id)}
													>删除</Button
												>
											</td>
										</tr>
										{#if child.children?.length}
											{#each child.children as sub}
												<tr
													class="border-b border-slate-100 bg-slate-50/30"
												>
													<td class="px-4 py-3 pl-12"
														>— — {sub.id}</td
													>
													<td class="px-4 py-3 pl-12"
														>— — {sub.name}</td
													>
													<td class="px-4 py-3"
														>{sub.slug || "-"}</td
													>
													<td class="px-4 py-3"
														>{sub.type ||
															"category"}</td
													>
													<td class="px-4 py-3"
														>{sub.sort}</td
													>
													<td class="px-4 py-3"
														>{sub.status === 1
															? "启用"
															: "禁用"}</td
													>
													<td
														class="px-4 py-3 text-right"
													>
														<Button
															variant="ghost"
															size="sm"
															onclick={() =>
																openCreate(
																	sub.id,
																)}
															>添加子级</Button
														>
														<Button
															variant="ghost"
															size="sm"
															onclick={() =>
																openEdit(sub)}
															>编辑</Button
														>
														<Button
															variant="ghost"
															size="sm"
															onclick={() =>
																remove(sub.id)}
															>删除</Button
														>
													</td>
												</tr>
											{/each}
										{/if}
									{/each}
								{/if}
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</CardContent>
	</Card>
</div>

{#if dialogOpen}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
		onclick={() => (dialogOpen = false)}
		onkeydown={(e) => e.key === "Escape" && (dialogOpen = false)}
		role="button"
		tabindex="-1"
	>
		<div
			class="w-full max-w-md rounded-lg bg-white p-6 shadow-xl"
			onclick={(e) => e.stopPropagation()}
			onkeydown={(e) => e.stopPropagation()}
			role="dialog"
		>
			<h3 class="mb-4 text-lg font-semibold">
				{editing ? "编辑分类" : "新增分类"}
			</h3>
			<div class="space-y-4">
				<div>
					<label class="mb-1 block text-sm font-medium">父级</label>
					<select
						bind:value={form.parent_id}
						class="h-9 w-full rounded-md border border-slate-200 px-3"
					>
						{#each parentOptions as opt}
							<option value={opt.id}
								>{"　".repeat(opt.indent)}{opt.name}</option
							>
						{/each}
					</select>
				</div>
				<div>
					<label class="mb-1 block text-sm font-medium">名称</label>
					<Input bind:value={form.name} placeholder="分类名称" />
				</div>
				<div>
					<label class="mb-1 block text-sm font-medium">Slug</label>
					<Input bind:value={form.slug} placeholder="url-slug" />
				</div>
				<div>
					<label class="mb-1 block text-sm font-medium">类型</label>
					<select
						bind:value={form.type}
						class="h-9 w-full rounded-md border px-3"
					>
						<option value="category">category</option>
						<option value="link">link</option>
					</select>
				</div>
				<div>
					<label class="mb-1 block text-sm font-medium">链接</label>
					<Input bind:value={form.link} placeholder="https://..." />
				</div>
				<div>
					<label class="mb-1 block text-sm font-medium">排序</label>
					<input
						type="number"
						bind:value={form.sort}
						class="flex h-9 w-full rounded-md border border-slate-200 px-3 py-1 text-sm"
					/>
				</div>
				<div>
					<label class="mb-1 block text-sm font-medium">状态</label>
					<select
						bind:value={form.status}
						class="h-9 w-full rounded-md border px-3"
					>
						<option value={1}>启用</option>
						<option value={0}>禁用</option>
					</select>
				</div>
			</div>
			<div class="mt-6 flex justify-end gap-2">
				<Button variant="outline" onclick={() => (dialogOpen = false)}
					>取消</Button
				>
				<Button onclick={save}>保存</Button>
			</div>
		</div>
	</div>
{/if}
