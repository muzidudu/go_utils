<script lang="ts">
	import { onMount } from "svelte";
	import { api, type Site, type CreateSiteReq } from "$lib/api";
	import { Button } from "$lib/components/ui/button";
	import { Card, CardHeader, CardContent } from "$lib/components/ui/card";
	import { Input } from "$lib/components/ui/input";
	import * as Select from "$lib/components/ui/select";

	let sites = $state<Site[]>([]);
	let templates = $state<string[]>([]);
	let loading = $state(true);
	let error = $state("");
	let dialogOpen = $state(false);
	let editing = $state<Site | null>(null);
	let form = $state<CreateSiteReq>({
		name: "",
		domain: "",
		subdomains: [],
		template: "default",
		is_default: false,
		status: 1,
	});
	let subdomainsStr = $state("");

	async function loadTemplates() {
		try {
			templates = await api.templates.list();
		} catch (e) {
			error = e instanceof Error ? e.message : "加载模板失败";
		} finally {
			loading = false;
		}
	}

	async function load() {
		loading = true;
		error = "";
		try {
			sites = await api.sites.list();
		} catch (e) {
			error = e instanceof Error ? e.message : "加载失败";
		} finally {
			loading = false;
		}
	}

	function openCreate() {
		editing = null;
		form = {
			name: "",
			domain: "",
			subdomains: [],
			template: "default",
			is_default: false,
			status: 1,
		};
		subdomainsStr = "";
		dialogOpen = true;
	}

	function openEdit(site: Site) {
		editing = site;
		form = {
			name: site.name,
			domain: site.domain,
			subdomains: [...(site.subdomains || [])],
			template: site.template || "default",
			is_default: site.is_default,
			status: site.status,
		};
		subdomainsStr = (site.subdomains || []).join(", ");
		dialogOpen = true;
	}

	async function save() {
		form.subdomains = subdomainsStr
			.split(",")
			.map((s) => s.trim())
			.filter(Boolean);
		try {
			if (editing) {
				await api.sites.update(editing.id, form);
			} else {
				await api.sites.create(form);
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
			await api.sites.delete(id);
			await load();
		} catch (e) {
			error = e instanceof Error ? e.message : "删除失败";
		}
	}

	onMount(async () => {
		await Promise.all([load(), loadTemplates()]);
	});
</script>

<svelte:head>
	<title>站点管理</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold text-slate-900">站点管理</h1>
		<Button onclick={openCreate}>新增站点</Button>
	</div>

	{#if error}
		<div class="rounded-md bg-red-50 p-4 text-sm text-red-700">{error}</div>
	{/if}

	<Card>
		<CardHeader>
			<h2 class="text-lg font-semibold">站点列表</h2>
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
									>ID</th
								>
								<th class="px-4 py-3 text-left font-medium"
									>名称</th
								>
								<th class="px-4 py-3 text-left font-medium"
									>域名</th
								>
								<th class="px-4 py-3 text-left font-medium"
									>模板</th
								>
								<th class="px-4 py-3 text-left font-medium"
									>默认</th
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
							{#each sites as site}
								<tr
									class="border-b border-slate-100 hover:bg-slate-50"
								>
									<td class="px-4 py-3">{site.id}</td>
									<td class="px-4 py-3">{site.name}</td>
									<td class="px-4 py-3">{site.domain}</td>
									<td class="px-4 py-3">{site.template}</td>
									<td class="px-4 py-3"
										>{site.is_default ? "是" : "否"}</td
									>
									<td class="px-4 py-3"
										>{site.status === 1
											? "启用"
											: "禁用"}</td
									>
									<td class="px-4 py-3 text-right">
										<Button
											variant="ghost"
											size="sm"
											onclick={() => openEdit(site)}
											>编辑</Button
										>
										<Button
											variant="ghost"
											size="sm"
											onclick={() => remove(site.id)}
											>删除</Button
										>
									</td>
								</tr>
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
		onkeydown={(e) => {
			if (e.key === "Escape") {
				dialogOpen = false;
			}
		}}
		role="button"
		tabindex="-1"
	>
		<div
			class="w-full max-w-md rounded-lg bg-white p-6 shadow-xl"
			onclick={(e) => {
				e.stopPropagation();
			}}
			onkeydown={(e) => {
				e.stopPropagation();
			}}
			role="dialog"
		>
			<h3 class="mb-4 text-lg font-semibold">
				{editing ? "编辑站点" : "新增站点"}
			</h3>
			<div class="space-y-4">
				<div>
					<label class="mb-1 block text-sm font-medium">名称</label>
					<Input bind:value={form.name} placeholder="站点名称" />
				</div>
				<div>
					<label class="mb-1 block text-sm font-medium">域名</label>
					<Input bind:value={form.domain} placeholder="example.com" />
				</div>
				<div>
					<label class="mb-1 block text-sm font-medium"
						>子域名（逗号分隔）</label
					>
					<Input
						bind:value={subdomainsStr}
						placeholder="www.example.com, m.example.com"
					/>
				</div>
				<div>
					<label class="mb-1 block text-sm font-medium">模板</label>

					<Select.Root type="single" bind:value={form.template}>
						<Select.Trigger>
							{form.template || "选择模板"}
						</Select.Trigger>
						<Select.Content>
							{#each templates as template}
								<Select.Item
									value={template}
									label={template}
								/>
							{/each}
						</Select.Content>
					</Select.Root>
				</div>
				<div class="flex items-center gap-2">
					<input
						type="checkbox"
						bind:checked={form.is_default}
						id="is_default"
					/>
					<label for="is_default" class="text-sm">默认站点</label>
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
