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

- `immediate_cast`：检测到满足条件时立刻释放的技能列表。仅允许 `ultimate` 和 `combo`。
- `sequential_cast`：依次释放的技能列表。允许 `skill`、`combo`、`ultimate`，但不得与 `immediate_cast` 有交集。

注意：如果 `sequential_cast` 中的条目类型是 `ultimate`，则必须同时提供 `window_ms`。

### 2.3 允许的扩展字段

为兼容阶段化、波次化、条件化排轴，允许在顺序释放项里保留这些扩展字段 (但并不必须)：

- `delay_ms`：相对前一项的延迟
- `window_ms`：可执行时间窗
- `allow_remedy`：允许补救（仅限终结技，值仅限 `1`）

对于字段“允许补救”，其含义是：若在实际战斗中，某技能在可执行时间窗内没有达到释放条件，则常规技能队列会将其舍弃转而执行轴的下一个步骤，而这个技能将单独压入一个特别的缓冲队列，暂时获得“亮了就点”属性，在后续战斗中仍然可以被释放，且一经释放则从缓冲队列里剔除，失去“亮了就点”属性.

注意到我们并不提供针对一般战技的“补救”选项，因为对于同一套阵容在实战中很少会出现技力条存在巨大波动差异的现象，因此在协议里添加对战技的补救措施通常并不会使轴变得更加稳定，反而可能会由于过于严格的时间窗限制导致程序出现更加严重的乱轴。因此在实际排轴工作中，我们**非常不建议您对战技设置严格的时间窗限制。**

同时，您还可能注意到我们同样不提供对连携技的“补救”选项，这是由于实战中可能会遇到“连携技激活但不释放，先开其他技能”的轴，且由于目前项目不支持对连携技对应干员的识别，添加“补救”同样大概率反而会导致更加意外的乱轴。不过游戏中连携技的触发机制提供了一种手动添加补救机制的措施：由于很大一部分连携技可以“在另外某个技能后触发”，因此您**可以在某些技能中间插入一个时间窗较短的 combo 项**，以对某些已知容易产生触发时机波动的连携技实行补救。

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
    "operator_count": 4,    // 由于目前项目只支持 4 人作战, 此处仅限填 4
    "immediate_cast": [
        {
            "type": "ultimate",
            "operator_id": 1
        },
        {
            "type": "combo"
        }
    ],
    "sequential_cast": [
        {
            "type": "ultimate",
            "operator_id": 2,
            "window_ms": 800,
            "allow_remedy": 1   // 并非必须
        },
        {
            "type": "skill",
            "operator_id": 3,
            "delay_ms": 1200,   // 并非必须
            "window_ms": 500    // 并非必须
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
