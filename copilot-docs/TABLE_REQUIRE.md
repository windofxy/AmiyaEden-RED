# 表格 UI 开发规范

> 基于项目现有实现（`useTable` + `ArtTable` + `ArtTableHeader`）归纳，新增表格页面须严格遵循本规范。

---

## 目录

1. [页面结构](#1-页面结构)
2. [核心 Hook：useTable](#2-核心-hookusетable)
3. [ArtTableHeader 组件](#3-arttableheader-组件)
4. [ArtTable 组件](#4-arttable-组件)
5. [列定义规范（columnsFactory）](#5-列定义规范columnsfactory)
6. [操作列与 ArtButtonTable](#6-操作列与-artbuttontable)
7. [状态/枚举渲染](#7-状态枚举渲染)
8. [搜索栏](#8-搜索栏)
9. [新增/编辑对话框](#9-新增编辑对话框)
10. [删除确认](#10-删除确认)
11. [权限/子对话框](#11-权限子对话框)
12. [原生 ElTable 使用场景](#12-原生-eltable-使用场景)
13. [API 函数命名约定](#13-api-函数命名约定)
14. [常见错误与禁止事项](#14-常见错误与禁止事项)

---

## 1. 页面结构

标准管理页面的 HTML 骨架如下：

```vue
<template>
  <div class="[页面名]-page art-full-height">
    <!-- 可选：搜索栏（独立组件，放在卡片外） -->
    <XxxSearch
      v-model="searchForm"
      @search="handleSearch"
      @reset="resetSearchParams"
    />

    <ElCard class="art-table-card" shadow="never">
      <!-- 表格头部（工具栏） -->
      <ArtTableHeader
        v-model:columns="columnChecks"
        :loading="loading"
        @refresh="refreshData"
      >
        <template #left>
          <!-- 左侧操作按钮 -->
          <ElButton type="primary" :icon="Plus" @click="openCreateDialog"
            >新增xxx</ElButton
          >
        </template>
      </ArtTableHeader>

      <!-- 表格主体 -->
      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <!-- CRUD 对话框（放在 ElCard 外） -->
    <ElDialog ... />

    <!-- 子功能对话框（放在 ElCard 外） -->
    <XxxDialog v-model:visible="subDialogVisible" ... />
  </div>
</template>
```

**关键点：**

| 条目         | 规范                                                      |
| ------------ | --------------------------------------------------------- |
| 根容器 class | `art-full-height`（确保表格高度自适应）                   |
| 卡片         | `ElCard` with `shadow="never"` + `class="art-table-card"` |
| 对话框位置   | 放在 `ElCard` **外部**，与卡片同级                        |
| 搜索栏位置   | 放在 `ElCard` **外部**，卡片之上                          |

---

## 2. 核心 Hook：useTable

所有带分页的管理页面**必须**使用 `useTable`，禁止手动管理 `loading`、`data`、`pagination`。

### 2.1 基础用法

```typescript
import { useTable } from "@/hooks/core/useTable";

const {
  columns, // 当前可见列配置（传给 ArtTable）
  columnChecks, // 列勾选状态（传给 ArtTableHeader v-model:columns）
  data, // 表格数据
  loading, // 加载状态
  pagination, // 分页信息 { current, size, total }
  getData, // 手动触发加载（带参数）
  searchParams, // 当前搜索参数对象（可直接 Object.assign 修改）
  resetSearchParams, // 重置搜索参数并刷新
  handleSizeChange, // 每页条数变化（传给 @pagination:size-change）
  handleCurrentChange, // 当前页变化（传给 @pagination:current-change）
  refreshData, // 无感刷新（保持当前页）
} = useTable({
  core: {
    apiFn: fetchGetXxxList, // API 函数，必填
    apiParams: { current: 1, size: 20 }, // 默认请求参数
    columnsFactory: () => [
      /* 列定义 */
    ],
  },
});
```

### 2.2 useTable 配置速查

```typescript
useTable({
  core: {
    apiFn, // (params) => Promise<PaginatedResponse<T>>  必填
    apiParams, // 初始请求参数，含 current / size
    columnsFactory, // () => ColumnOption[]  必填（需要列管理时）
    immediate, // 是否立即请求，默认 true
    excludeParams, // 排除传给后端的字段名数组
  },
  transform: {
    dataTransformer, // 数据转换函数
    responseAdapter, // 响应适配器（非标准格式时使用）
  },
  performance: {
    enableCache, // 是否开启缓存，默认 false
    cacheTime, // 缓存时间（ms），默认 5min
  },
});
```

### 2.3 刷新策略选择

| 场景       | 使用方法                             |
| ---------- | ------------------------------------ |
| 新增记录后 | `refreshData()` —— 保持当前页        |
| 编辑记录后 | `refreshData()` —— 保持当前页        |
| 删除记录后 | `refreshData()`                      |
| 搜索/重置  | `getData()` 或 `resetSearchParams()` |

---

## 3. ArtTableHeader 组件

工具栏，提供刷新、尺寸调整、全屏、列设置、其他设置功能。

### 3.1 必传 Props

```vue
<ArtTableHeader
  v-model:columns="columnChecks"   <!-- 列勾选双向绑定，必填 -->
  :loading="loading"               <!-- 加载状态，用于刷新图标动画 -->
  @refresh="refreshData"           <!-- 手动刷新回调 -->
>
  <template #left>
    <!-- 新增、批量导入等主操作按钮放在 #left 插槽 -->
  </template>
</ArtTableHeader>
```

### 3.2 可选 Props

| Prop                   | 类型      | 默认值                                              | 说明                                  |
| ---------------------- | --------- | --------------------------------------------------- | ------------------------------------- |
| `layout`               | `string`  | `'search,refresh,size,fullscreen,columns,settings'` | 控制显示的工具按钮                    |
| `showZebra`            | `boolean` | `true`                                              | 显示斑马纹开关                        |
| `showBorder`           | `boolean` | `true`                                              | 显示边框开关                          |
| `showHeaderBackground` | `boolean` | `true`                                              | 显示表头背景开关                      |
| `showSearchBar`        | `boolean` | `undefined`                                         | 搜索栏显示状态（配合 `@search` 事件） |

### 3.3 插槽

| 插槽     | 说明                               |
| -------- | ---------------------------------- |
| `#left`  | 左侧操作区，放「新增」等主操作按钮 |
| `#right` | 右侧工具区末尾，放自定义工具图标   |

---

## 4. ArtTable 组件

表格主体，封装了 `ElTable` + 分页器。

### 4.1 标准用法

```vue
<ArtTable
  :loading="loading"
  :data="data"
  :columns="columns"
  :pagination="pagination"
  @pagination:size-change="handleSizeChange"
  @pagination:current-change="handleCurrentChange"
/>
```

### 4.2 Props 速查

| Prop         | 类型                       | 说明                                            |
| ------------ | -------------------------- | ----------------------------------------------- |
| `loading`    | `boolean`                  | 加载骨架屏                                      |
| `data`       | `any[]`                    | 表格行数据                                      |
| `columns`    | `ColumnOption[]`           | 列配置（来自 `useTable` 返回的 `columns`）      |
| `pagination` | `{ current, size, total }` | 分页状态（来自 `useTable` 返回的 `pagination`） |
| `emptyText`  | `string`                   | 空数据提示，默认「暂无数据」                    |
| `stripe`     | `boolean`                  | 斑马纹（不填则由全局设置控制）                  |
| `border`     | `boolean`                  | 边框（不填则由全局设置控制）                    |

### 4.3 事件

| 事件                         | 说明                                 |
| ---------------------------- | ------------------------------------ |
| `@pagination:size-change`    | 每页条数变化 → 传 `handleSizeChange` |
| `@pagination:current-change` | 页码变化 → 传 `handleCurrentChange`  |

> **不要**把 `el-table` 原生事件直接绑定在 `ArtTable` 外层，`ArtTable` 通过 `$attrs` 透传，直接写即可。

---

## 5. 列定义规范（columnsFactory）

`columnsFactory` 是一个**返回列配置数组的工厂函数**，在 `useTable` 的 `core.columnsFactory` 中声明。

### 5.1 固定列顺序

```
[序号列] → [业务数据列...] → [操作列（fixed: 'right'）]
```

### 5.2 序号列

```typescript
{ type: 'index', width: 60, label: '序号' }
```

全局索引（跨页连续）用 `type: 'globalIndex'`。

### 5.3 普通数据列

```typescript
{
  prop: 'name',          // 对应数据字段名，必填
  label: '名称',          // 列头文字，必填
  width: 160,            // 固定宽度（px），文字短的列建议设置
  minWidth: 200,         // 最小宽度（弹性列）
  showOverflowTooltip: true,  // 文字超出时 tooltip 显示，文本列必须加
  sortable: true,        // 是否支持排序
  fixed: 'left' | 'right'    // 固定列
}
```

**宽度建议：**

| 内容类型      | 建议 width                             |
| ------------- | -------------------------------------- |
| 序号          | 60                                     |
| 短标签（Tag） | 80–120                                 |
| 名称/标题     | 140–180                                |
| 描述/备注     | `minWidth: 200`（弹性）                |
| 时间          | 180                                    |
| IP            | 140                                    |
| 操作列        | 根据按钮数量：1钮=80，2钮=120，3钮=200 |

### 5.4 自定义渲染列（formatter）

使用 `formatter` 返回 VNode（`h` 函数），**禁止**在 `formatter` 中直接使用模板字符串拼接 HTML。

```typescript
{
  prop: 'status',
  label: '状态',
  width: 100,
  formatter: (row: XxxItem) =>
    h(ElTag, { type: row.status === 1 ? 'success' : 'danger', size: 'small' }, () =>
      row.status === 1 ? '正常' : '禁用'
    )
}
```

### 5.5 操作列

```typescript
{
  prop: 'actions',
  label: '操作',
  width: 200,
  fixed: 'right',   // 操作列必须固定在右侧
  formatter: (row: XxxItem) =>
    h('div', { class: 'flex gap-1' }, [
      // 按钮列表
    ])
}
```

---

## 6. 操作列与 ArtButtonTable

操作列按钮**统一使用 `ArtButtonTable`**，禁止手写 `ElButton` 小圆钮。

### 6.1 内置类型

| `type`     | 图标   | 样式   | 用途         |
| ---------- | ------ | ------ | ------------ |
| `'add'`    | 加号   | 主题色 | 新增（行内） |
| `'edit'`   | 铅笔   | 橙色   | 编辑         |
| `'delete'` | 垃圾桶 | 红色   | 删除         |
| `'view'`   | 眼睛   | 蓝色   | 查看详情     |
| `'more'`   | 三点   | 无色   | 更多操作     |

### 6.2 标准操作列示例

```typescript
import ArtButtonTable from "@/components/core/forms/art-button-table/index.vue";
import { Setting } from "@element-plus/icons-vue";
import { ElButton } from "element-plus";

// 在 formatter 中：
h("div", { class: "flex gap-1" }, [
  // 自定义功能按钮（用 ElButton small）
  h(
    ElButton,
    {
      size: "small",
      type: "warning",
      icon: Setting,
      onClick: () => openXxx(row),
    },
    () => "权限",
  ),

  // 标准编辑
  h(ArtButtonTable, { type: "edit", onClick: () => openEditDialog(row) }),

  // 标准删除（系统项禁用）
  h(ArtButtonTable, {
    type: "delete",
    disabled: row.is_system,
    onClick: () => handleDelete(row),
  }),
]);
```

**规则：**

- 按钮顺序：**自定义功能 → 编辑 → 删除**
- 系统内置记录的删除按钮必须加 `:disabled="row.is_system"`
- 操作列容器 class 用 `flex gap-1`

---

## 7. 状态/枚举渲染

### 7.1 使用配置映射表（禁止用 if-else 链）

```typescript
// 在 <script setup> 顶部定义
const STATUS_CONFIG: Record<number, { type: string; text: string }> = {
  1: { type: "success", text: "正常" },
  0: { type: "danger", text: "禁用" },
};

const ROLE_CONFIG: Record<string, { type: string; text: string }> = {
  super_admin: { type: "danger", text: "超级管理员" },
  admin: { type: "warning", text: "管理员" },
  user: { type: "success", text: "已认证用户" },
  guest: { type: "info", text: "访客" },
};
```

### 7.2 布尔值（是/否）渲染

```typescript
formatter: (row) =>
  h(ElTag, { type: row.is_system ? "danger" : "info", size: "small" }, () =>
    row.is_system ? "是" : "否",
  );
```

### 7.3 Tag 尺寸

列表中的 Tag 统一使用 `size: 'small'`。

### 7.4 系统角色/特殊角色区分

使用 `effect: 'dark'` 来突出显示系统级别的角色：

```typescript
h(
  ElTag,
  {
    type: CODE_TYPE[row.code] as any,
    effect: row.is_system ? "dark" : "plain",
    size: "small",
  },
  () => row.code,
);
```

---

## 8. 搜索栏

搜索栏**抽离为独立组件**（`modules/xxx-search.vue`），不内联在主页面。

### 8.1 搜索组件接口约定

```typescript
// xxx-search.vue
const props = defineProps<{ modelValue: XxxSearchForm }>();
const emit = defineEmits<{
  (e: "update:modelValue", v: XxxSearchForm): void;
  (e: "search", params: XxxSearchForm): void;
  (e: "reset"): void;
}>();
```

### 8.2 主页面集成

```vue
<XxxSearch
  v-model="searchForm"
  @search="handleSearch"
  @reset="resetSearchParams"
/>
```

```typescript
const searchForm = ref({ field1: undefined, field2: undefined });

const handleSearch = (params: Record<string, any>) => {
  Object.assign(searchParams, params);
  getData();
};
```

---

## 9. 新增/编辑对话框

### 9.1 对话框结构

```vue
<ElDialog
  v-model="dialogVisible"
  :title="editingItem ? '编辑xxx' : '新增xxx'"
  width="480px"
  destroy-on-close
>
  <ElForm ref="formRef" :model="formData" :rules="formRules" label-width="80px">
    <!-- 编辑时 code/主键 字段禁用 -->
    <ElFormItem label="编码" prop="code">
      <ElInput v-model="formData.code" :disabled="!!editingItem" />
    </ElFormItem>
    <ElFormItem label="名称" prop="name">
      <ElInput v-model="formData.name" />
    </ElFormItem>
  </ElForm>

  <template #footer>
    <ElButton @click="dialogVisible = false">取消</ElButton>
    <ElButton type="primary" :loading="submitLoading" @click="handleSubmit">确定</ElButton>
  </template>
</ElDialog>
```

### 9.2 表单状态管理

```typescript
const dialogVisible = ref(false)
const submitLoading = ref(false)
const formRef = ref<FormInstance>()
const editingItem = ref<XxxItem | null>(null)   // null = 新增模式

const formData = reactive({
  code: '',
  name: '',
  description: '',
  sort: 0
})

function resetForm() {
  formData.code = ''
  formData.name = ''
  formData.description = ''
  formData.sort = 0
  editingItem.value = null
}

function openCreateDialog() {
  resetForm()
  dialogVisible.value = true
}

function openEditDialog(row: XxxItem) {
  editingItem.value = row
  // 填充 formData
  Object.assign(formData, { code: row.code, name: row.name, ... })
  dialogVisible.value = true
}
```

### 9.3 提交逻辑

```typescript
async function handleSubmit() {
  if (!formRef.value) return;
  await formRef.value.validate(); // 校验失败自动抛出，不需要手动 catch
  submitLoading.value = true;
  try {
    if (editingItem.value) {
      await fetchUpdateXxx(editingItem.value.id, {
        /* 仅可修改字段 */
      });
      ElMessage.success("更新成功");
    } else {
      await fetchCreateXxx({ ...formData });
      ElMessage.success("创建成功");
    }
    dialogVisible.value = false;
    refreshData();
  } catch (e: any) {
    ElMessage.error(e?.message ?? "操作失败");
  } finally {
    submitLoading.value = false;
  }
}
```

### 9.4 表单校验规则约定

```typescript
const formRules: FormRules = {
  code: [
    { required: true, message: "请输入编码", trigger: "blur" },
    {
      pattern: /^[a-z][a-z0-9_]*$/,
      message: "小写字母开头，仅含字母/数字/下划线",
      trigger: "blur",
    },
  ],
  name: [{ required: true, message: "请输入名称", trigger: "blur" }],
};
```

---

## 10. 删除确认

删除**必须**使用 `ElMessageBox.confirm`，禁止直接调用删除接口。

```typescript
async function handleDelete(row: XxxItem) {
  if (row.is_system) return; // 系统内置项直接拦截

  await ElMessageBox.confirm(`确定要删除「${row.name}」吗？`, "确认删除", {
    confirmButtonText: "删除",
    cancelButtonText: "取消",
    type: "warning",
  });
  // 用户点取消会 throw，不会继续执行
  try {
    await fetchDeleteXxx(row.id);
    ElMessage.success("删除成功");
    refreshData();
  } catch (e: any) {
    ElMessage.error(e?.message ?? "删除失败");
  }
}
```

---

## 11. 权限/子对话框

子功能对话框（如权限分配、用户角色分配）**抽离为独立组件**。

### 11.1 接口约定

```typescript
// role-permission-dialog.vue
defineProps<{
  visible: boolean;
  roleId?: number;
  roleName?: string;
}>();
defineEmits<{
  (e: "update:visible", v: boolean): void;
  (e: "saved"): void; // 保存成功后通知父页面刷新
}>();
```

### 11.2 父页面调用

```typescript
const permVisible = ref(false);
const permRoleId = ref<number>();
const permRoleName = ref("");

function openPermDialog(row: RoleItem) {
  permRoleId.value = row.id;
  permRoleName.value = row.name;
  permVisible.value = true;
}
```

```vue
<RolePermissionDialog
  v-model:visible="permVisible"
  :role-id="permRoleId"
  :role-name="permRoleName"
  @saved="refreshData"
/>
```

---

## 12. 原生 ElTable 使用场景

以下场景**可以**不使用 `useTable`，直接使用原生 `ElTable`：

- 无分页的静态/小数据量展示表格
- 排行榜等特殊布局（使用 `ElCard #header` 插槽自定义头部）
- 数据通过其他方式管理（如 Pinia store）

```vue
<!-- PAP 排行榜示例 -->
<ElCard class="art-table-card" shadow="never">
  <template #header>
    <div class="flex items-center justify-between flex-wrap gap-3">
      <!-- 自定义 header -->
    </div>
  </template>

  <ElTable v-loading="loading" :data="list" stripe border>
    <ElTableColumn type="index" width="55" label="排名" align="center" />
    <ElTableColumn prop="name" label="名称" min-width="140" />
    <!-- ... -->
  </ElTable>
</ElCard>
```

---

## 13. API 函数命名约定

在 `src/api/` 下的模块文件中，函数命名遵循以下格式：

```
fetch + 动词 + 模块名 + [细分]
```

| 操作         | 命名示例                                 |
| ------------ | ---------------------------------------- |
| 列表（分页） | `fetchGetRoleList`                       |
| 列表（全量） | `fetchGetAllRoles`                       |
| 单条         | `fetchGetRole`                           |
| 新增         | `fetchCreateRole`                        |
| 更新         | `fetchUpdateRole`                        |
| 删除         | `fetchDeleteRole`                        |
| 特殊子资源   | `fetchGetUserRoles`、`fetchSetUserRoles` |

---

## 14. 常见错误与禁止事项

| 禁止                                                       | 正确做法                                              |
| ---------------------------------------------------------- | ----------------------------------------------------- |
| 手动维护 `loading` / `data` / `pagination`                 | 使用 `useTable`                                       |
| 在 `formatter` 中用 `innerHTML` / 模板字符串拼接 HTML      | 用 `h()` 函数返回 VNode                               |
| 操作列按钮用小尺寸 `ElButton`                              | 使用 `ArtButtonTable`                                 |
| `ElCard` 不加 `shadow="never"` 和 `class="art-table-card"` | 统一加上                                              |
| 对话框不加 `destroy-on-close`                              | 必须加，防止表单状态残留                              |
| 系统内置项不做删除限制                                     | 判断 `row.is_system`，禁用按钮并在函数内 early return |
| 删除不经确认直接调用 API                                   | 必须先 `ElMessageBox.confirm`                         |
| 子对话框逻辑写在主页面                                     | 抽离为 `modules/xxx-dialog.vue`                       |
| 搜索栏写在主页面内                                         | 抽离为 `modules/xxx-search.vue`                       |
| 忘记给长文本列加 `showOverflowTooltip`                     | 描述/备注等长文本列必须加                             |
| 操作列不加 `fixed: 'right'`                                | 操作列必须固定右侧                                    |
| 页面直接显示文字                                           | 必须使用i18n                                          |
