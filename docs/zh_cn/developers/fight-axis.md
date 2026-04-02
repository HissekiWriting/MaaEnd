# 战斗排轴协议

## 1. 目标

战斗排轴协议用于描述自动战斗中“什么技能、由谁释放、在什么约束下释放”。它只负责策略描述，不负责点击、按键、识别或流程控制。

该协议的植入方式与 Pipeline 资源类似：

- 配置文件放在 `tools/schema/` 下，作为机器可校验的协议定义。
- 具体执行仍由 `assets/resource/pipeline/AutoFight/` 和 `agent/go-service/autofight/` 负责。
- 任务层可以通过 option / pipeline_override 选择不同排轴方案，也可以在 AutoFight 子选项里直接填写排轴 JSON 文件路径。

## 2. 协议边界

### 2.1 对象类型

协议仅允许以下技能类型：

- `skill`：一般战技
- `combo`：连携技
- `ultimate`：终结技

角色对象 id 必须是 `1` 到 `4` 的正整数，对应战斗内四名干员的槽位。

### 2.2 内容板块

协议分为两类内容板块：

- `immediate_cast`：检测到满足条件时立刻释放的技能列表。仅允许 `ultimate`。
- `sequential_cast`：依次释放的技能列表。允许 `skill`、`combo`、`ultimate`，但不得与 `immediate_cast` 有交集。

### 2.3 推荐的扩展字段

为兼容后续的阶段化、波次化、条件化排轴，建议在顺序释放项里保留这些扩展字段 (但并不必须)：

- `delay_ms`：相对前一项的延迟。
- `window_ms`：可执行时间窗。
- `priority`：同一时刻的内部优先级。
- `trigger`：任意扩展触发条件对象。

## 3. Schema 植入方式

仓库内已有统一的 schema 目录：`tools/schema/`。

新增排轴协议时，建议同步创建：

- `tools/schema/fight_axis.schema.json`：排轴协议 schema。
- `docs/zh_cn/developers/fight-axis.md`：协议说明文档。

这样做的好处是：

- 可以用 JSON Schema 做静态校验。
- 可以和现有 `pipeline.schema.json`、`interface.schema.json` 保持一致的工程结构。
- 后续如果要在任务中引用排轴文件，可以直接按 `pipeline_override` 风格接入。

## 3.1 任务 UI 接入建议

在 `assets/tasks/RealTimeTask.json` 中，建议将 `AutoFight` 选项展开为两个子项：

- `AutoFightAttack`：控制是否保留普攻锚点。
- `AutoFightAxisFile`：输入排轴 JSON 文件路径。

其中 `AutoFightAxisFile` 推荐使用 `input` 类型，默认值可设为 `assets/resource/pipeline/AutoFightAxis/default.json`。这样任务界面可以直接让用户自选排轴文件，同时保持未来接入执行层时的输入格式稳定。

## 4. 推荐的示例结构

```jsonc
{
    "version": 1,
    "name": "ExampleAxis",
    "description": "示例排轴",
    "battle_mode": "realtime",
    "operator_count": 4,
    "immediate_cast": [
        {
            "type": "ultimate",
            "operator_id": 1
        }
    ],
    "sequential_cast": [
        {
            "type": "combo",
            "operator_id": 2,
            "priority": 100     # 并非必须
        },
        {
            "type": "skill",
            "operator_id": 3,
            "delay_ms": 1200,   # 并非必须
            "window_ms": 500    # 并非必须
        }
    ]
}
```

## 5. 与现有 AutoFight 的关系

现有 AutoFight 的实现是“画面识别 -> 动作入队 -> Pipeline 执行”。新的排轴协议应当只作为“动作入队规则”的输入，不应把整套战斗流程迁移到 schema 之外的业务代码里。

也就是说：

- 识别是否可放技能，仍然走 `agent/go-service/autofight/`。
- 点击 / 按键执行，仍然走 `assets/resource/pipeline/AutoFight/Action.json`。
- 排轴文件只决定“该优先谁、何时释放、哪些条目互斥”。

## 6. 后续建议

如果要继续推进落地，下一步建议补两件事：

1. 给排轴 schema 增加示例文件，作为测试和文档样本。
2. 在 `AutoFight` 的任务 option 中增加一个“选择排轴文件”的输入项，让任务层可以切换不同阵容或关卡策略。
