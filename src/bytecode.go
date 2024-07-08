package main

import (
	"encoding/binary"
	"encoding/gob"
	"math"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

type StateType int32

const (
	ST_S StateType = 1 << iota
	ST_C
	ST_A
	ST_L
	ST_N
	ST_U
	ST_MASK = 1<<iota - 1
	ST_D    = ST_L
	ST_F    = ST_N
	ST_P    = ST_U
	ST_SCA  = ST_S | ST_C | ST_A
)

type AttackType int32

const (
	AT_NA AttackType = 1 << (iota + 6)
	AT_NT
	AT_NP
	AT_SA
	AT_ST
	AT_SP
	AT_HA
	AT_HT
	AT_HP
	AT_AA  = AT_NA | AT_SA | AT_HA
	AT_AT  = AT_NT | AT_ST | AT_HT
	AT_AP  = AT_NP | AT_SP | AT_HP
	AT_ALL = AT_AA | AT_AT | AT_AP
	AT_AN  = AT_NA | AT_NT | AT_NP
	AT_AS  = AT_SA | AT_ST | AT_SP
	AT_AH  = AT_HA | AT_HT | AT_HP
)

type MoveType int32

const (
	MT_I MoveType = 1 << (iota + 15)
	MT_H
	MT_A
	MT_U
	MT_MNS = MT_I
	MT_PLS = MT_H
)

type ValueType int

const (
	VT_None ValueType = iota
	VT_Float
	VT_Int
	VT_Bool
	VT_SFalse
)

type OpCode byte

const (
	OC_var OpCode = iota
	OC_sysvar
	OC_fvar
	OC_sysfvar
	OC_localvar
	OC_int8
	OC_int
	OC_int64
	OC_float
	OC_pop
	OC_dup
	OC_swap
	OC_run
	OC_nordrun
	OC_jsf8
	OC_jmp8
	OC_jz8
	OC_jnz8
	OC_jmp
	OC_jz
	OC_jnz
	OC_eq
	OC_ne
	OC_gt
	OC_le
	OC_lt
	OC_ge
	OC_neg
	OC_blnot
	OC_bland
	OC_blxor
	OC_blor
	OC_not
	OC_and
	OC_xor
	OC_or
	OC_add
	OC_sub
	OC_mul
	OC_div
	OC_mod
	OC_pow
	OC_abs
	OC_exp
	OC_ln
	OC_log
	OC_cos
	OC_sin
	OC_tan
	OC_acos
	OC_asin
	OC_atan
	OC_floor
	OC_ceil
	OC_ifelse
	OC_time
	OC_animtime
	OC_animelemtime
	OC_animelemno
	OC_statetype
	OC_movetype
	OC_ctrl
	OC_command
	OC_random
	OC_pos_x
	OC_pos_y
	OC_vel_x
	OC_vel_y
	OC_screenpos_x
	OC_screenpos_y
	OC_facing
	OC_anim
	OC_animexist
	OC_selfanimexist
	OC_alive
	OC_life
	OC_lifemax
	OC_power
	OC_powermax
	OC_canrecover
	OC_roundstate
	OC_ishelper
	OC_numhelper
	OC_numexplod
	OC_numprojid
	OC_numproj
	OC_teammode
	OC_teamside
	OC_hitdefattr
	OC_inguarddist
	OC_movecontact
	OC_movehit
	OC_moveguarded
	OC_movereversed
	OC_projcontacttime
	OC_projhittime
	OC_projguardedtime
	OC_projcanceltime
	OC_backedge
	OC_backedgedist
	OC_backedgebodydist
	OC_frontedge
	OC_frontedgedist
	OC_frontedgebodydist
	OC_leftedge
	OC_rightedge
	OC_topedge
	OC_bottomedge
	OC_camerapos_x
	OC_camerapos_y
	OC_camerazoom
	OC_gamewidth
	OC_gameheight
	OC_screenwidth
	OC_screenheight
	OC_stateno
	OC_prevstateno
	OC_id
	OC_playeridexist
	OC_gametime
	OC_numtarget
	OC_numenemy
	OC_numpartner
	OC_ailevel
	OC_palno
	OC_hitcount
	OC_uniqhitcount
	OC_hitpausetime
	OC_hitover
	OC_hitshakeover
	OC_hitfall
	OC_hitvel_x
	OC_hitvel_y
	OC_player
	OC_parent
	OC_root
	OC_helper
	OC_target
	OC_partner
	OC_enemy
	OC_enemynear
	OC_playerid
	OC_playerindex
	OC_helperindex
	OC_p2
	OC_stateowner
	OC_rdreset
	OC_const_
	OC_st_
	OC_ex_
	OC_ex2_
)
const (
	OC_const_data_life OpCode = iota
	OC_const_data_power
	OC_const_data_guardpoints
	OC_const_data_dizzypoints
	OC_const_data_attack
	OC_const_data_defence
	OC_const_data_fall_defence_up
	OC_const_data_fall_defence_mul
	OC_const_data_liedown_time
	OC_const_data_airjuggle
	OC_const_data_sparkno
	OC_const_data_guard_sparkno
	OC_const_data_hitsound_channel
	OC_const_data_guardsound_channel
	OC_const_data_ko_echo
	OC_const_data_intpersistindex
	OC_const_data_floatpersistindex
	OC_const_size_xscale
	OC_const_size_yscale
	OC_const_size_ground_back
	OC_const_size_ground_front
	OC_const_size_air_back
	OC_const_size_air_front
	OC_const_size_height_stand
	OC_const_size_height_crouch
	OC_const_size_height_air_top
	OC_const_size_height_air_bottom
	OC_const_size_height_down
	OC_const_size_attack_dist_front
	OC_const_size_attack_dist_back
	OC_const_size_attack_z_width_back
	OC_const_size_attack_z_width_front
	OC_const_size_proj_attack_dist_front
	OC_const_size_proj_attack_dist_back
	OC_const_size_proj_doscale
	OC_const_size_head_pos_x
	OC_const_size_head_pos_y
	OC_const_size_mid_pos_x
	OC_const_size_mid_pos_y
	OC_const_size_shadowoffset
	OC_const_size_draw_offset_x
	OC_const_size_draw_offset_y
	OC_const_size_z_width
	OC_const_size_z_enable
	OC_const_velocity_walk_fwd_x
	OC_const_velocity_walk_back_x
	OC_const_velocity_walk_up_x
	OC_const_velocity_walk_down_x
	OC_const_velocity_run_fwd_x
	OC_const_velocity_run_fwd_y
	OC_const_velocity_run_back_x
	OC_const_velocity_run_back_y
	OC_const_velocity_run_up_x
	OC_const_velocity_run_up_y
	OC_const_velocity_run_down_x
	OC_const_velocity_run_down_y
	OC_const_velocity_jump_y
	OC_const_velocity_jump_neu_x
	OC_const_velocity_jump_back_x
	OC_const_velocity_jump_fwd_x
	OC_const_velocity_jump_up_x
	OC_const_velocity_jump_down_x
	OC_const_velocity_runjump_back_x
	OC_const_velocity_runjump_back_y
	OC_const_velocity_runjump_y
	OC_const_velocity_runjump_fwd_x
	OC_const_velocity_runjump_up_x
	OC_const_velocity_runjump_down_x
	OC_const_velocity_airjump_y
	OC_const_velocity_airjump_neu_x
	OC_const_velocity_airjump_back_x
	OC_const_velocity_airjump_fwd_x
	OC_const_velocity_airjump_up_x
	OC_const_velocity_airjump_down_x
	OC_const_velocity_air_gethit_groundrecover_x
	OC_const_velocity_air_gethit_groundrecover_y
	OC_const_velocity_air_gethit_airrecover_mul_x
	OC_const_velocity_air_gethit_airrecover_mul_y
	OC_const_velocity_air_gethit_airrecover_add_x
	OC_const_velocity_air_gethit_airrecover_add_y
	OC_const_velocity_air_gethit_airrecover_back
	OC_const_velocity_air_gethit_airrecover_fwd
	OC_const_velocity_air_gethit_airrecover_up
	OC_const_velocity_air_gethit_airrecover_down
	OC_const_velocity_air_gethit_ko_add_x
	OC_const_velocity_air_gethit_ko_add_y
	OC_const_velocity_air_gethit_ko_ymin
	OC_const_velocity_ground_gethit_ko_xmul
	OC_const_velocity_ground_gethit_ko_add_x
	OC_const_velocity_ground_gethit_ko_add_y
	OC_const_velocity_ground_gethit_ko_ymin
	OC_const_movement_airjump_num
	OC_const_movement_airjump_height
	OC_const_movement_yaccel
	OC_const_movement_stand_friction
	OC_const_movement_crouch_friction
	OC_const_movement_stand_friction_threshold
	OC_const_movement_crouch_friction_threshold
	OC_const_movement_air_gethit_groundlevel
	OC_const_movement_air_gethit_groundrecover_ground_threshold
	OC_const_movement_air_gethit_groundrecover_groundlevel
	OC_const_movement_air_gethit_airrecover_threshold
	OC_const_movement_air_gethit_airrecover_yaccel
	OC_const_movement_air_gethit_trip_groundlevel
	OC_const_movement_down_bounce_offset_x
	OC_const_movement_down_bounce_offset_y
	OC_const_movement_down_bounce_yaccel
	OC_const_movement_down_bounce_groundlevel
	OC_const_movement_down_friction_threshold
	OC_const_name
	OC_const_p2name
	OC_const_p3name
	OC_const_p4name
	OC_const_p5name
	OC_const_p6name
	OC_const_p7name
	OC_const_p8name
	OC_const_authorname
	OC_const_displayname
	OC_const_stagevar_info_author
	OC_const_stagevar_info_displayname
	OC_const_stagevar_info_name
	OC_const_stagevar_camera_boundleft
	OC_const_stagevar_camera_boundright
	OC_const_stagevar_camera_boundhigh
	OC_const_stagevar_camera_boundlow
	OC_const_stagevar_camera_verticalfollow
	OC_const_stagevar_camera_floortension
	OC_const_stagevar_camera_tensionhigh
	OC_const_stagevar_camera_tensionlow
	OC_const_stagevar_camera_tension
	OC_const_stagevar_camera_tensionvel
	OC_const_stagevar_camera_cuthigh
	OC_const_stagevar_camera_cutlow
	OC_const_stagevar_camera_startzoom
	OC_const_stagevar_camera_zoomout
	OC_const_stagevar_camera_zoomin
	OC_const_stagevar_camera_zoomindelay
	OC_const_stagevar_camera_ytension_enable
	OC_const_stagevar_camera_autocenter
	OC_const_stagevar_playerinfo_leftbound
	OC_const_stagevar_playerinfo_rightbound
	OC_const_stagevar_scaling_topscale
	OC_const_stagevar_bound_screenleft
	OC_const_stagevar_bound_screenright
	OC_const_stagevar_stageinfo_localcoord_x
	OC_const_stagevar_stageinfo_localcoord_y
	OC_const_stagevar_stageinfo_xscale
	OC_const_stagevar_stageinfo_yscale
	OC_const_stagevar_stageinfo_zoffset
	OC_const_stagevar_stageinfo_zoffsetlink
	OC_const_stagevar_shadow_intensity
	OC_const_stagevar_shadow_color_r
	OC_const_stagevar_shadow_color_g
	OC_const_stagevar_shadow_color_b
	OC_const_stagevar_shadow_yscale
	OC_const_stagevar_shadow_fade_range_begin
	OC_const_stagevar_shadow_fade_range_end
	OC_const_stagevar_shadow_xshear
	OC_const_stagevar_reflection_intensity
	OC_const_constants
	OC_const_stage_constants
)
const (
	OC_st_var OpCode = iota
	OC_st_sysvar
	OC_st_fvar
	OC_st_sysfvar
	OC_st_varadd
	OC_st_sysvaradd
	OC_st_fvaradd
	OC_st_sysfvaradd
	OC_st_map
)
const (
	OC_ex_p2dist_x OpCode = iota
	OC_ex_p2dist_y
	OC_ex_p2bodydist_x
	OC_ex_p2bodydist_y
	OC_ex_parentdist_x
	OC_ex_parentdist_y
	OC_ex_rootdist_x
	OC_ex_rootdist_y
	OC_ex_win
	OC_ex_winko
	OC_ex_wintime
	OC_ex_winperfect
	OC_ex_winspecial
	OC_ex_winhyper
	OC_ex_lose
	OC_ex_loseko
	OC_ex_losetime
	OC_ex_drawgame
	OC_ex_matchover
	OC_ex_matchno
	OC_ex_roundno
	OC_ex_roundsexisted
	OC_ex_ishometeam
	OC_ex_tickspersecond
	OC_ex_majorversion
	OC_ex_drawpalno
	OC_ex_const240p
	OC_ex_const480p
	OC_ex_const720p
	OC_ex_const1080p
	OC_ex_gethitvar_animtype
	OC_ex_gethitvar_air_animtype
	OC_ex_gethitvar_ground_animtype
	OC_ex_gethitvar_fall_animtype
	OC_ex_gethitvar_type
	OC_ex_gethitvar_airtype
	OC_ex_gethitvar_groundtype
	OC_ex_gethitvar_damage
	OC_ex_gethitvar_hitcount
	OC_ex_gethitvar_fallcount
	OC_ex_gethitvar_hitshaketime
	OC_ex_gethitvar_hittime
	OC_ex_gethitvar_slidetime
	OC_ex_gethitvar_ctrltime
	OC_ex_gethitvar_recovertime
	OC_ex_gethitvar_xoff
	OC_ex_gethitvar_yoff
	OC_ex_gethitvar_xvel
	OC_ex_gethitvar_yvel
	OC_ex_gethitvar_yaccel
	OC_ex_gethitvar_chainid
	OC_ex_gethitvar_guarded
	OC_ex_gethitvar_isbound
	OC_ex_gethitvar_fall
	OC_ex_gethitvar_fall_damage
	OC_ex_gethitvar_fall_xvel
	OC_ex_gethitvar_fall_yvel
	OC_ex_gethitvar_fall_recover
	OC_ex_gethitvar_fall_time
	OC_ex_gethitvar_fall_recovertime
	OC_ex_gethitvar_fall_kill
	OC_ex_gethitvar_fall_envshake_time
	OC_ex_gethitvar_fall_envshake_freq
	OC_ex_gethitvar_fall_envshake_ampl
	OC_ex_gethitvar_fall_envshake_phase
	OC_ex_gethitvar_fall_envshake_mul
	OC_ex_gethitvar_attr
	OC_ex_gethitvar_dizzypoints
	OC_ex_gethitvar_guardpoints
	OC_ex_gethitvar_id
	OC_ex_gethitvar_playerno
	OC_ex_gethitvar_redlife
	OC_ex_gethitvar_score
	OC_ex_gethitvar_hitdamage
	OC_ex_gethitvar_guarddamage
	OC_ex_gethitvar_power
	OC_ex_gethitvar_hitpower
	OC_ex_gethitvar_guardpower
	OC_ex_gethitvar_kill
	OC_ex_gethitvar_priority
	OC_ex_gethitvar_guardcount
	OC_ex_gethitvar_facing
	OC_ex_gethitvar_ground_velocity_x
	OC_ex_gethitvar_ground_velocity_y
	OC_ex_gethitvar_air_velocity_x
	OC_ex_gethitvar_air_velocity_y
	OC_ex_gethitvar_down_velocity_x
	OC_ex_gethitvar_down_velocity_y
	OC_ex_gethitvar_guard_velocity_x
	OC_ex_gethitvar_airguard_velocity_x
	OC_ex_gethitvar_airguard_velocity_y
	OC_ex_gethitvar_frame
	OC_ex_ailevelf
	OC_ex_animelemlength
	OC_ex_animframe_alphadest
	OC_ex_animframe_angle
	OC_ex_animframe_alphasource
	OC_ex_animframe_group
	OC_ex_animframe_hflip
	OC_ex_animframe_image
	OC_ex_animframe_time
	OC_ex_animframe_vflip
	OC_ex_animframe_xoffset
	OC_ex_animframe_xscale
	OC_ex_animframe_yoffset
	OC_ex_animframe_yscale
	OC_ex_animframe_numclsn1
	OC_ex_animframe_numclsn2
	OC_ex_animlength
	OC_ex_attack
	OC_ex_combocount
	OC_ex_consecutivewins
	OC_ex_defence
	OC_ex_dizzy
	OC_ex_dizzypoints
	OC_ex_dizzypointsmax
	OC_ex_fighttime
	OC_ex_firstattack
	OC_ex_framespercount
	OC_ex_float
	OC_ex_gamemode
	OC_ex_getplayerid
	OC_ex_groundangle
	OC_ex_guardbreak
	OC_ex_guardpoints
	OC_ex_guardpointsmax
	OC_ex_helperid
	OC_ex_helperindexexist
	OC_ex_helpername
	OC_ex_hitoverridden
	OC_ex_inputtime_B
	OC_ex_inputtime_D
	OC_ex_inputtime_F
	OC_ex_inputtime_U
	OC_ex_inputtime_L
	OC_ex_inputtime_R
	OC_ex_inputtime_a
	OC_ex_inputtime_b
	OC_ex_inputtime_c
	OC_ex_inputtime_x
	OC_ex_inputtime_y
	OC_ex_inputtime_z
	OC_ex_inputtime_s
	OC_ex_inputtime_d
	OC_ex_inputtime_w
	OC_ex_inputtime_m
	OC_ex_movehitvar_frame
	OC_ex_movehitvar_cornerpush
	OC_ex_movehitvar_id
	OC_ex_movehitvar_overridden
	OC_ex_movehitvar_playerno
	OC_ex_movehitvar_spark_x
	OC_ex_movehitvar_spark_y
	OC_ex_movehitvar_uniqhit
	OC_ex_incustomstate
	OC_ex_indialogue
	OC_ex_isassertedchar
	OC_ex_isassertedglobal
	OC_ex_ishost
	OC_ex_jugglepoints
	OC_ex_localcoord_x
	OC_ex_localcoord_y
	OC_ex_localscale
	OC_ex_maparray
	OC_ex_max
	OC_ex_min
	OC_ex_numplayer
	OC_ex_clamp
	OC_ex_sign
	OC_ex_atan2
	OC_ex_rad
	OC_ex_deg
	OC_ex_lastplayerid
	OC_ex_lerp
	OC_ex_memberno
	OC_ex_movecountered
	OC_ex_mugenversion
	OC_ex_pausetime
	OC_ex_physics
	OC_ex_playerno
	OC_ex_playerindexexist
	OC_ex_randomrange
	OC_ex_ratiolevel
	OC_ex_receiveddamage
	OC_ex_receivedhits
	OC_ex_redlife
	OC_ex_round
	OC_ex_roundtype
	OC_ex_score
	OC_ex_scoretotal
	OC_ex_selfstatenoexist
	OC_ex_sprpriority
	OC_ex_stagebackedgedist
	OC_ex_stagefrontedgedist
	OC_ex_stagetime
	OC_ex_standby
	OC_ex_teamleader
	OC_ex_teamsize
	OC_ex_timeelapsed
	OC_ex_timeremaining
	OC_ex_timetotal
	OC_ex_playercount
	OC_ex_pos_z
	OC_ex_vel_z
	OC_ex_prevanim
	OC_ex_prevmovetype
	OC_ex_prevstatetype
	OC_ex_reversaldefattr
	OC_ex_bgmlength
	OC_ex_bgmposition
	OC_ex_airjumpcount
	OC_ex_envshakevar_time
	OC_ex_envshakevar_freq
	OC_ex_envshakevar_ampl
	OC_ex_angle
	OC_ex_scale_x
	OC_ex_scale_y
	OC_ex_offset_x
	OC_ex_offset_y
	OC_ex_alpha_s
	OC_ex_alpha_d
	OC_ex_selfcommand
	OC_ex_guardcount
	OC_ex_gamefps
	OC_ex_fightscreenvar_info_author
	OC_ex_fightscreenvar_info_name
	OC_ex_fightscreenvar_round_ctrl_time
	OC_ex_fightscreenvar_round_over_hittime
	OC_ex_fightscreenvar_round_over_time
	OC_ex_fightscreenvar_round_over_waittime
	OC_ex_fightscreenvar_round_over_wintime
	OC_ex_fightscreenvar_round_slow_time
	OC_ex_fightscreenvar_round_start_waittime
)
const (
	OC_ex2_index OpCode = iota
	OC_ex2_runorder
	OC_ex2_palfxvar_time
	OC_ex2_palfxvar_addr
	OC_ex2_palfxvar_addg
	OC_ex2_palfxvar_addb
	OC_ex2_palfxvar_mulr
	OC_ex2_palfxvar_mulg
	OC_ex2_palfxvar_mulb
	OC_ex2_palfxvar_color
	OC_ex2_palfxvar_hue
	OC_ex2_palfxvar_invertall
	OC_ex2_palfxvar_invertblend
	OC_ex2_palfxvar_bg_time
	OC_ex2_palfxvar_bg_addr
	OC_ex2_palfxvar_bg_addg
	OC_ex2_palfxvar_bg_addb
	OC_ex2_palfxvar_bg_mulr
	OC_ex2_palfxvar_bg_mulg
	OC_ex2_palfxvar_bg_mulb
	OC_ex2_palfxvar_bg_color
	OC_ex2_palfxvar_bg_hue
	OC_ex2_palfxvar_bg_invertall
	OC_ex2_palfxvar_all_time
	OC_ex2_palfxvar_all_addr
	OC_ex2_palfxvar_all_addg
	OC_ex2_palfxvar_all_addb
	OC_ex2_palfxvar_all_mulr
	OC_ex2_palfxvar_all_mulg
	OC_ex2_palfxvar_all_mulb
	OC_ex2_palfxvar_all_color
	OC_ex2_palfxvar_all_hue
	OC_ex2_palfxvar_all_invertall
	OC_ex2_palfxvar_all_invertblend
	OC_ex2_introstate
	OC_ex2_bgmvar_filename
	OC_ex2_bgmvar_loopend
	OC_ex2_bgmvar_loopstart
	OC_ex2_bgmvar_startposition
	OC_ex2_bgmvar_volume
	OC_ex2_gameoption_sound_panningrange
	OC_ex2_gameoption_sound_wavchannels
	OC_ex2_gameoption_sound_mastervolume
	OC_ex2_gameoption_sound_wavvolume
	OC_ex2_gameoption_sound_bgmvolume
	OC_ex2_gameoption_sound_maxvolume
	OC_ex2_groundlevel
)
const (
	NumVar     = 60
	NumSysVar  = 5
	NumFvar    = 40
	NumSysFvar = 5
)

type StringPool struct {
	List []string
	Map  map[string]int
}

func NewStringPool() *StringPool {
	return &StringPool{Map: make(map[string]int)}
}
func (sp *StringPool) Clear() {
	sp.List, sp.Map = nil, make(map[string]int)
}
func (sp *StringPool) Add(s string) int {
	i, ok := sp.Map[s]
	if !ok {
		i = len(sp.List)
		sp.List = append(sp.List, s)
		sp.Map[s] = i
	}
	return i
}

type BytecodeValue struct {
	t ValueType
	v float64
}

func (bv BytecodeValue) IsNone() bool { return bv.t == VT_None }
func (bv BytecodeValue) IsSF() bool   { return bv.t == VT_SFalse }
func (bv BytecodeValue) ToF() float32 {
	if bv.IsSF() {
		return 0
	}
	return float32(bv.v)
}
func (bv BytecodeValue) ToI() int32 {
	if bv.IsSF() {
		return 0
	}
	return int32(bv.v)
}
func (bv BytecodeValue) ToI64() int64 {
	if bv.IsSF() {
		return 0
	}
	return int64(bv.v)
}
func (bv BytecodeValue) ToB() bool {
	if bv.IsSF() || bv.v == 0 {
		return false
	}
	return true
}
func (bv *BytecodeValue) SetF(f float32) {
	if math.IsNaN(float64(f)) {
		*bv = BytecodeSF()
	} else {
		*bv = BytecodeValue{VT_Float, float64(f)}
	}
}
func (bv *BytecodeValue) SetI(i int32) {
	*bv = BytecodeValue{VT_Int, float64(i)}
}
func (bv *BytecodeValue) SetI64(i int64) {
	*bv = BytecodeValue{VT_Int, float64(i)}
}
func (bv *BytecodeValue) SetB(b bool) {
	bv.t = VT_Bool
	bv.v = float64(Btoi(b))
}

func bvNone() BytecodeValue {
	return BytecodeValue{VT_None, 0}
}
func BytecodeSF() BytecodeValue {
	return BytecodeValue{VT_SFalse, math.NaN()}
}
func BytecodeFloat(f float32) BytecodeValue {
	return BytecodeValue{VT_Float, float64(f)}
}
func BytecodeInt(i int32) BytecodeValue {
	return BytecodeValue{VT_Int, float64(i)}
}
func BytecodeInt64(i int64) BytecodeValue {
	return BytecodeValue{VT_Int, float64(i)}
}
func BytecodeBool(b bool) BytecodeValue {
	return BytecodeValue{VT_Bool, float64(Btoi(b))}
}

type BytecodeStack []BytecodeValue

func (bs *BytecodeStack) Clear()                { *bs = (*bs)[:0] }
func (bs *BytecodeStack) Push(bv BytecodeValue) { *bs = append(*bs, bv) }
func (bs *BytecodeStack) PushI(i int32)         { bs.Push(BytecodeInt(i)) }
func (bs *BytecodeStack) PushI64(i int64)       { bs.Push(BytecodeInt64(i)) }
func (bs *BytecodeStack) PushF(f float32)       { bs.Push(BytecodeFloat(f)) }
func (bs *BytecodeStack) PushB(b bool)          { bs.Push(BytecodeBool(b)) }
func (bs BytecodeStack) Top() *BytecodeValue {
	return &bs[len(bs)-1]
}
func (bs *BytecodeStack) Pop() (bv BytecodeValue) {
	bv, *bs = *bs.Top(), (*bs)[:len(*bs)-1]
	return
}
func (bs *BytecodeStack) Dup() {
	bs.Push(*bs.Top())
}
func (bs *BytecodeStack) Swap() {
	*bs.Top(), (*bs)[len(*bs)-2] = (*bs)[len(*bs)-2], *bs.Top()
}
func (bs *BytecodeStack) Alloc(size int) []BytecodeValue {
	if len(*bs)+size > cap(*bs) {
		tmp := *bs
		*bs = make(BytecodeStack, len(*bs)+size)
		copy(*bs, tmp)
	} else {
		*bs = (*bs)[:len(*bs)+size]
		for i := len(*bs) - size; i < len(*bs); i++ {
			(*bs)[i] = bvNone()
		}
	}
	return (*bs)[len(*bs)-size:]
}

type BytecodeExp []OpCode

func Float32frombytes(bytes []byte) float32 {
	bits := binary.LittleEndian.Uint32(bytes)
	float := math.Float32frombits(bits)
	return float
}

func (be *BytecodeExp) append(op ...OpCode) {
	*be = append(*be, op...)
}
func (be *BytecodeExp) appendValue(bv BytecodeValue) (ok bool) {
	switch bv.t {
	case VT_Float:
		be.append(OC_float)
		f := float32(bv.v)
		be.append((*(*[4]OpCode)(unsafe.Pointer(&f)))[:]...)
	case VT_Int:
		if bv.v >= -128 && bv.v <= 127 {
			be.append(OC_int8, OpCode(bv.v))
		} else if bv.v >= math.MinInt32 && bv.v <= math.MaxInt32 {
			be.append(OC_int)
			i := int32(bv.v)
			be.append((*(*[4]OpCode)(unsafe.Pointer(&i)))[:]...)
		} else {
			be.append(OC_int64)
			i := int64(bv.v)
			be.append((*(*[8]OpCode)(unsafe.Pointer(&i)))[:]...)
		}
	case VT_Bool:
		if bv.v != 0 {
			be.append(OC_int8, 1)
		} else {
			be.append(OC_int8, 0)
		}
	case VT_SFalse:
		be.append(OC_int8, 0)
	default:
		return false
	}
	return true
}
func (be *BytecodeExp) appendI32Op(op OpCode, addr int32) {
	be.append(op)
	be.append((*(*[4]OpCode)(unsafe.Pointer(&addr)))[:]...)
}
func (be *BytecodeExp) appendI64Op(op OpCode, addr int64) {
	be.append(op)
	be.append((*(*[8]OpCode)(unsafe.Pointer(&addr)))[:]...)
}
func (BytecodeExp) neg(v *BytecodeValue) {
	if v.t == VT_Float {
		v.v *= -1
	} else {
		v.SetI(-v.ToI())
	}
}
func (BytecodeExp) not(v *BytecodeValue) {
	v.SetI(^v.ToI())
}
func (BytecodeExp) blnot(v *BytecodeValue) {
	v.SetB(!v.ToB())
}
func (BytecodeExp) pow(v1 *BytecodeValue, v2 BytecodeValue, pn int) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetF(Pow(v1.ToF(), v2.ToF()))
	} else if v2.ToF() < 0 {
		v1.SetF(Pow(v1.ToF(), v2.ToF()))
	} else {
		i1, i2, hb := v1.ToI(), v2.ToI(), int32(-1)
		for uint32(i2)>>uint(hb+1) != 0 {
			hb++
		}
		var i, bit, tmp int32 = 1, 0, i1
		for ; bit <= hb; bit++ {
			var shift uint
			if bit == hb || sys.cgi[pn].mugenver[0] == 1 {
				shift = uint(bit)
			} else {
				shift = uint((hb - 1) - bit)
			}
			if i2&(1<<shift) != 0 {
				i *= tmp
			}
			tmp *= tmp
		}
		v1.SetI(i)
	}
}
func (BytecodeExp) mul(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetF(v1.ToF() * v2.ToF())
	} else {
		v1.SetI(v1.ToI() * v2.ToI())
	}
}
func (BytecodeExp) div(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetF(v1.ToF() / v2.ToF())
	} else if v2.ToI() == 0 {
		*v1 = BytecodeSF()
	} else {
		v1.SetI(v1.ToI() / v2.ToI())
	}
}
func (BytecodeExp) mod(v1 *BytecodeValue, v2 BytecodeValue) {
	if v2.ToI() == 0 {
		*v1 = BytecodeSF()
	} else {
		v1.SetI(v1.ToI() % v2.ToI())
	}
}
func (BytecodeExp) add(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetF(v1.ToF() + v2.ToF())
	} else {
		v1.SetI(v1.ToI() + v2.ToI())
	}
}
func (BytecodeExp) sub(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetF(v1.ToF() - v2.ToF())
	} else {
		v1.SetI(v1.ToI() - v2.ToI())
	}
}
func (BytecodeExp) gt(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetB(v1.ToF() > v2.ToF())
	} else {
		v1.SetB(v1.ToI() > v2.ToI())
	}
}
func (BytecodeExp) ge(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetB(v1.ToF() >= v2.ToF())
	} else {
		v1.SetB(v1.ToI() >= v2.ToI())
	}
}
func (BytecodeExp) lt(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetB(v1.ToF() < v2.ToF())
	} else {
		v1.SetB(v1.ToI() < v2.ToI())
	}
}
func (BytecodeExp) le(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetB(v1.ToF() <= v2.ToF())
	} else {
		v1.SetB(v1.ToI() <= v2.ToI())
	}
}
func (BytecodeExp) eq(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetB(v1.ToF() == v2.ToF())
	} else {
		v1.SetB(v1.ToI() == v2.ToI())
	}
}
func (BytecodeExp) ne(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetB(v1.ToF() != v2.ToF())
	} else {
		v1.SetB(v1.ToI() != v2.ToI())
	}
}
func (BytecodeExp) and(v1 *BytecodeValue, v2 BytecodeValue) {
	v1.SetI(v1.ToI() & v2.ToI())
}
func (BytecodeExp) xor(v1 *BytecodeValue, v2 BytecodeValue) {
	v1.SetI(v1.ToI() ^ v2.ToI())
}
func (BytecodeExp) or(v1 *BytecodeValue, v2 BytecodeValue) {
	v1.SetI(v1.ToI() | v2.ToI())
}
func (BytecodeExp) bland(v1 *BytecodeValue, v2 BytecodeValue) {
	v1.SetB(v1.ToB() && v2.ToB())
}
func (BytecodeExp) blxor(v1 *BytecodeValue, v2 BytecodeValue) {
	v1.SetB(v1.ToB() != v2.ToB())
}
func (BytecodeExp) blor(v1 *BytecodeValue, v2 BytecodeValue) {
	v1.SetB(v1.ToB() || v2.ToB())
}
func (BytecodeExp) abs(v1 *BytecodeValue) {
	if v1.t == VT_Float {
		v1.v = math.Abs(v1.v)
	} else {
		v1.SetI(Abs(v1.ToI()))
	}
}
func (BytecodeExp) exp(v1 *BytecodeValue) {
	v1.SetF(float32(math.Exp(v1.v)))
}
func (BytecodeExp) ln(v1 *BytecodeValue) {
	if v1.v <= 0 {
		*v1 = BytecodeSF()
	} else {
		v1.SetF(float32(math.Log(v1.v)))
	}
}
func (BytecodeExp) log(v1 *BytecodeValue, v2 BytecodeValue) {
	if v1.v <= 0 || v2.v <= 0 {
		*v1 = BytecodeSF()
	} else {
		v1.SetF(float32(math.Log(v2.v) / math.Log(v1.v)))
	}
}
func (BytecodeExp) cos(v1 *BytecodeValue) {
	v1.SetF(float32(math.Cos(v1.v)))
}
func (BytecodeExp) sin(v1 *BytecodeValue) {
	v1.SetF(float32(math.Sin(v1.v)))
}
func (BytecodeExp) tan(v1 *BytecodeValue) {
	v1.SetF(float32(math.Tan(v1.v)))
}
func (BytecodeExp) acos(v1 *BytecodeValue) {
	v1.SetF(float32(math.Acos(v1.v)))
}
func (BytecodeExp) asin(v1 *BytecodeValue) {
	v1.SetF(float32(math.Asin(v1.v)))
}
func (BytecodeExp) atan(v1 *BytecodeValue) {
	v1.SetF(float32(math.Atan(v1.v)))
}
func (BytecodeExp) floor(v1 *BytecodeValue) {
	if v1.t == VT_Float {
		f := math.Floor(v1.v)
		if math.IsNaN(f) {
			*v1 = BytecodeSF()
		} else {
			v1.SetI(int32(f))
		}
	}
}
func (BytecodeExp) ceil(v1 *BytecodeValue) {
	if v1.t == VT_Float {
		f := math.Ceil(v1.v)
		if math.IsNaN(f) {
			*v1 = BytecodeSF()
		} else {
			v1.SetI(int32(f))
		}
	}
}
func (BytecodeExp) max(v1 *BytecodeValue, v2 BytecodeValue) {
	if v1.v >= v2.v {
		v1.SetF(float32(v1.v))
	} else {
		v1.SetF(float32(v2.v))
	}
}
func (BytecodeExp) min(v1 *BytecodeValue, v2 BytecodeValue) {
	if v1.v <= v2.v {
		v1.SetF(float32(v1.v))
	} else {
		v1.SetF(float32(v2.v))
	}
}
func (BytecodeExp) random(v1 *BytecodeValue, v2 BytecodeValue) {
	v1.SetI(RandI(int32(v1.v), int32(v2.v)))
}
func (BytecodeExp) round(v1 *BytecodeValue, v2 BytecodeValue) {
	shift := math.Pow(10, v2.v)
	v1.SetF(float32(math.Floor((v1.v*shift)+0.5) / shift))
}
func (BytecodeExp) clamp(v1 *BytecodeValue, v2 BytecodeValue, v3 BytecodeValue) {
	if v1.v <= v2.v {
		v1.SetF(float32(v2.v))
	} else if v1.v >= v3.v {
		v1.SetF(float32(v3.v))
	} else {
		v1.SetF(float32(v1.v))
	}
}
func (BytecodeExp) atan2(v1 *BytecodeValue, v2 BytecodeValue) {
	v1.SetF(float32(math.Atan2(v1.v, v2.v)))
}
func (BytecodeExp) sign(v1 *BytecodeValue) {
	if v1.v < 0 {
		v1.SetI(int32(-1))
	} else if v1.v > 0 {
		v1.SetI(int32(1))
	} else {
		v1.SetI(int32(0))
	}
}
func (BytecodeExp) rad(v1 *BytecodeValue) {
	v1.SetF(float32(v1.v * math.Pi / 180))
}
func (BytecodeExp) deg(v1 *BytecodeValue) {
	v1.SetF(float32(v1.v * 180 / math.Pi))
}
func (BytecodeExp) lerp(v1 *BytecodeValue, v2 BytecodeValue, v3 BytecodeValue) {
	amount := v3.v
	if v3.v <= 0 {
		amount = 0
	} else if v3.v >= 1 {
		amount = 1
	}
	v1.SetF(float32(v1.v + (v2.v-v1.v)*amount))
}
func (be BytecodeExp) run(c *Char) BytecodeValue {
	oc := c
	for i := 1; i <= len(be); i++ {
		switch be[i-1] {
		case OC_jsf8:
			if sys.bcStack.Top().IsSF() {
				if be[i] == 0 {
					i = len(be)
				} else {
					i += int(uint8(be[i])) + 1
				}
			} else {
				i++
			}
		case OC_jz8, OC_jnz8:
			if sys.bcStack.Top().ToB() == (be[i-1] == OC_jz8) {
				i++
				break
			}
			fallthrough
		case OC_jmp8:
			if be[i] == 0 {
				i = len(be)
			} else {
				i += int(uint8(be[i])) + 1
			}
		case OC_jz, OC_jnz:
			if sys.bcStack.Top().ToB() == (be[i-1] == OC_jz) {
				i += 4
				break
			}
			fallthrough
		case OC_jmp:
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_player:
			if c = sys.playerID(c.getPlayerID(int(sys.bcStack.Pop().ToI()))); c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_parent:
			if c = c.parent(); c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_root:
			if c = c.root(); c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_helper:
			if c = c.helper(sys.bcStack.Pop().ToI()); c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_target:
			if c = c.target(sys.bcStack.Pop().ToI()); c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_partner:
			if c = c.partner(sys.bcStack.Pop().ToI(), true); c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_enemy:
			if c = c.enemy(sys.bcStack.Pop().ToI()); c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_enemynear:
			if c = c.enemyNear(sys.bcStack.Pop().ToI()); c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_playerid:
			if c = sys.playerID(sys.bcStack.Pop().ToI()); c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_playerindex:
			if c = sys.playerIndex(sys.bcStack.Pop().ToI()); c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_p2:
			if c = c.p2(); c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_stateowner:
			if c = sys.chars[c.ss.sb.playerNo][0]; c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_helperindex:
			if c = c.helperByIndex(sys.bcStack.Pop().ToI()); c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_rdreset:
			// NOP
		case OC_run:
			l := int(*(*int32)(unsafe.Pointer(&be[i])))
			sys.bcStack.Push(be[i+4 : i+4+l].run(c))
			i += 4 + l
		case OC_nordrun:
			l := int(*(*int32)(unsafe.Pointer(&be[i])))
			sys.bcStack.Push(be[i+4 : i+4+l].run(oc))
			i += 4 + l
			continue
		case OC_int8:
			sys.bcStack.PushI(int32(int8(be[i])))
			i++
		case OC_int:
			sys.bcStack.PushI(*(*int32)(unsafe.Pointer(&be[i])))
			i += 4
		case OC_int64:
			sys.bcStack.PushI64(*(*int64)(unsafe.Pointer(&be[i])))
			i += 8
		case OC_float:
			arr := make([]byte, 4)
			arr[0] = byte(be[i])
			arr[1] = byte(be[i+1])
			arr[2] = byte(be[i+2])
			arr[3] = byte(be[i+3])
			flo := Float32frombytes(arr)
			sys.bcStack.PushF(flo)
			i += 4
		case OC_neg:
			be.neg(sys.bcStack.Top())
		case OC_not:
			be.not(sys.bcStack.Top())
		case OC_blnot:
			be.blnot(sys.bcStack.Top())
		case OC_pow:
			v2 := sys.bcStack.Pop()
			be.pow(sys.bcStack.Top(), v2, sys.workingChar.ss.sb.playerNo)
		case OC_mul:
			v2 := sys.bcStack.Pop()
			be.mul(sys.bcStack.Top(), v2)
		case OC_div:
			v2 := sys.bcStack.Pop()
			be.div(sys.bcStack.Top(), v2)
		case OC_mod:
			v2 := sys.bcStack.Pop()
			be.mod(sys.bcStack.Top(), v2)
		case OC_add:
			v2 := sys.bcStack.Pop()
			be.add(sys.bcStack.Top(), v2)
		case OC_sub:
			v2 := sys.bcStack.Pop()
			be.sub(sys.bcStack.Top(), v2)
		case OC_gt:
			v2 := sys.bcStack.Pop()
			be.gt(sys.bcStack.Top(), v2)
		case OC_ge:
			v2 := sys.bcStack.Pop()
			be.ge(sys.bcStack.Top(), v2)
		case OC_lt:
			v2 := sys.bcStack.Pop()
			be.lt(sys.bcStack.Top(), v2)
		case OC_le:
			v2 := sys.bcStack.Pop()
			be.le(sys.bcStack.Top(), v2)
		case OC_eq:
			v2 := sys.bcStack.Pop()
			be.eq(sys.bcStack.Top(), v2)
		case OC_ne:
			v2 := sys.bcStack.Pop()
			be.ne(sys.bcStack.Top(), v2)
		case OC_and:
			v2 := sys.bcStack.Pop()
			be.and(sys.bcStack.Top(), v2)
		case OC_xor:
			v2 := sys.bcStack.Pop()
			be.xor(sys.bcStack.Top(), v2)
		case OC_or:
			v2 := sys.bcStack.Pop()
			be.or(sys.bcStack.Top(), v2)
		case OC_bland:
			v2 := sys.bcStack.Pop()
			be.bland(sys.bcStack.Top(), v2)
		case OC_blxor:
			v2 := sys.bcStack.Pop()
			be.blxor(sys.bcStack.Top(), v2)
		case OC_blor:
			v2 := sys.bcStack.Pop()
			be.blor(sys.bcStack.Top(), v2)
		case OC_abs:
			be.abs(sys.bcStack.Top())
		case OC_exp:
			be.exp(sys.bcStack.Top())
		case OC_ln:
			be.ln(sys.bcStack.Top())
		case OC_log:
			v2 := sys.bcStack.Pop()
			be.log(sys.bcStack.Top(), v2)
		case OC_cos:
			be.cos(sys.bcStack.Top())
		case OC_sin:
			be.sin(sys.bcStack.Top())
		case OC_tan:
			be.tan(sys.bcStack.Top())
		case OC_acos:
			be.acos(sys.bcStack.Top())
		case OC_asin:
			be.asin(sys.bcStack.Top())
		case OC_atan:
			be.atan(sys.bcStack.Top())
		case OC_floor:
			be.floor(sys.bcStack.Top())
		case OC_ceil:
			be.ceil(sys.bcStack.Top())
		case OC_ifelse:
			v3 := sys.bcStack.Pop()
			v2 := sys.bcStack.Pop()
			if sys.bcStack.Top().ToB() {
				*sys.bcStack.Top() = v2
			} else {
				*sys.bcStack.Top() = v3
			}
		case OC_pop:
			sys.bcStack.Pop()
		case OC_dup:
			sys.bcStack.Dup()
		case OC_swap:
			sys.bcStack.Swap()
		case OC_ailevel:
			if !c.asf(ASF_noailevel) {
				sys.bcStack.PushI(int32(c.aiLevel()))
			} else {
				sys.bcStack.PushI(0)
			}
		case OC_alive:
			sys.bcStack.PushB(c.alive())
		case OC_anim:
			sys.bcStack.PushI(c.animNo)
		case OC_animelemno:
			*sys.bcStack.Top() = c.animElemNo(sys.bcStack.Top().ToI())
		case OC_animelemtime:
			*sys.bcStack.Top() = c.animElemTime(sys.bcStack.Top().ToI())
		case OC_animexist:
			*sys.bcStack.Top() = c.animExist(sys.workingChar, *sys.bcStack.Top())
		case OC_animtime:
			sys.bcStack.PushI(c.animTime())
		case OC_backedge:
			sys.bcStack.PushF(c.backEdge() * (c.localscl / oc.localscl))
		case OC_backedgebodydist:
			sys.bcStack.PushI(int32(c.backEdgeBodyDist() * (c.localscl / oc.localscl)))
		case OC_backedgedist:
			sys.bcStack.PushI(int32(c.backEdgeDist() * (c.localscl / oc.localscl)))
		case OC_bottomedge:
			sys.bcStack.PushF(c.bottomEdge() * (c.localscl / oc.localscl))
		case OC_camerapos_x:
			sys.bcStack.PushF(sys.cam.Pos[0] / oc.localscl)
		case OC_camerapos_y:
			sys.bcStack.PushF((sys.cam.Pos[1] + sys.cam.aspectcorrection + sys.cam.zoomanchorcorrection) / oc.localscl)
		case OC_camerazoom:
			sys.bcStack.PushF(sys.cam.Scale)
		case OC_canrecover:
			sys.bcStack.PushB(c.canRecover())
		case OC_command:
			if c.cmd == nil {
				sys.bcStack.PushB(false)
			} else {
				cmdName := sys.stringPool[sys.workingState.playerNo].List[*(*int32)(unsafe.Pointer(&be[i]))]
				redir := c.playerNo
				pno := c.playerNo
				// For a Mugen character, the command position is checked in the redirecting char
				// Recovery command is an exception in that its position is always checked in the final char
				if cmdName != "recovery" && oc.stWgi().ikemenver[0] == 0 && oc.stWgi().ikemenver[1] == 0 {
					redir = oc.ss.sb.playerNo
					pno = c.ss.sb.playerNo
				}
				cmdPos, ok := c.cmd[redir].Names[cmdName]
				ok = ok && c.command(pno, cmdPos)
				sys.bcStack.PushB(ok)
			}
			i += 4
		case OC_ctrl:
			sys.bcStack.PushB(c.ctrl())
		case OC_facing:
			sys.bcStack.PushI(int32(c.facing))
		case OC_frontedge:
			sys.bcStack.PushF(c.frontEdge() * (c.localscl / oc.localscl))
		case OC_frontedgebodydist:
			sys.bcStack.PushI(int32(c.frontEdgeBodyDist() * (c.localscl / oc.localscl)))
		case OC_frontedgedist:
			sys.bcStack.PushI(int32(c.frontEdgeDist() * (c.localscl / oc.localscl)))
		case OC_gameheight:
			// Optional exception preventing GameHeight from being affected by stage zoom.
			if c.stWgi().mugenver[0] == 1 && c.stWgi().mugenver[1] == 0 &&
				c.gi().constants["default.legacygamedistancespec"] == 1 {
				sys.bcStack.PushF(c.screenHeight())
			} else {
				sys.bcStack.PushF(c.gameHeight())
			}
		case OC_gametime:
			var pfTime int32
			if sys.netInput != nil {
				pfTime = sys.netInput.preFightTime
			} else if sys.fileInput != nil {
				pfTime = sys.fileInput.pfTime
			} else {
				pfTime = sys.preFightTime
			}
			sys.bcStack.PushI(sys.gameTime + pfTime)
		case OC_gamewidth:
			// Optional exception preventing GameWidth from being affected by stage zoom.
			if c.stWgi().mugenver[0] == 1 && c.stWgi().mugenver[1] == 0 &&
				c.gi().constants["default.legacygamedistancespec"] == 1 {
				sys.bcStack.PushF(c.screenWidth())
			} else {
				sys.bcStack.PushF(c.gameWidth())
			}
		case OC_hitcount:
			sys.bcStack.PushI(c.hitCount)
		case OC_hitdefattr:
			sys.bcStack.PushB(c.hitDefAttr(*(*int32)(unsafe.Pointer(&be[i]))))
			i += 4
		case OC_hitfall:
			sys.bcStack.PushB(c.ghv.fallf)
		case OC_hitover:
			sys.bcStack.PushB(c.hitOver())
		case OC_hitpausetime:
			sys.bcStack.PushI(c.hitPauseTime)
		case OC_hitshakeover:
			sys.bcStack.PushB(c.hitShakeOver())
		case OC_hitvel_x:
			sys.bcStack.PushF(c.hitVelX() * (c.localscl / oc.localscl))
		case OC_hitvel_y:
			sys.bcStack.PushF(c.hitVelY() * (c.localscl / oc.localscl))
		case OC_id:
			sys.bcStack.PushI(c.id)
		case OC_inguarddist:
			sys.bcStack.PushB(c.inguarddist)
		case OC_ishelper:
			*sys.bcStack.Top() = c.isHelper(*sys.bcStack.Top())
		case OC_leftedge:
			sys.bcStack.PushF(c.leftEdge() * (c.localscl / oc.localscl))
		case OC_life:
			sys.bcStack.PushI(c.life)
		case OC_lifemax:
			sys.bcStack.PushI(c.lifeMax)
		case OC_movecontact:
			sys.bcStack.PushI(c.moveContact())
		case OC_moveguarded:
			sys.bcStack.PushI(c.moveGuarded())
		case OC_movehit:
			sys.bcStack.PushI(c.moveHit())
		case OC_movereversed:
			sys.bcStack.PushI(c.moveReversed())
		case OC_movetype:
			sys.bcStack.PushB(c.ss.moveType == MoveType(be[i])<<15)
			i++
		case OC_numenemy:
			sys.bcStack.PushI(c.numEnemy())
		case OC_numexplod:
			*sys.bcStack.Top() = c.numExplod(*sys.bcStack.Top())
		case OC_numhelper:
			*sys.bcStack.Top() = c.numHelper(*sys.bcStack.Top())
		case OC_numpartner:
			sys.bcStack.PushI(c.numPartner())
		case OC_numproj:
			sys.bcStack.PushI(c.numProj())
		case OC_numprojid:
			*sys.bcStack.Top() = c.numProjID(*sys.bcStack.Top())
		case OC_numtarget:
			*sys.bcStack.Top() = c.numTarget(*sys.bcStack.Top())
		case OC_palno:
			sys.bcStack.PushI(c.palno())
		case OC_pos_x:
			var bindVelx float32
			if c.bindToId > 0 && !math.IsNaN(float64(c.bindPos[0])) && c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
				if sys.playerID(c.bindToId) != nil {
					bindVelx = c.vel[0]
				}
			}
			sys.bcStack.PushF(((c.pos[0]+bindVelx)*(c.localscl/oc.localscl) - sys.cam.Pos[0]/oc.localscl))
		case OC_pos_y:
			var bindVely float32
			if c.bindToId > 0 && !math.IsNaN(float64(c.bindPos[1])) && c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
				if sys.playerID(c.bindToId) != nil {
					bindVely = c.vel[1]
				}
			}
			sys.bcStack.PushF((c.pos[1] + bindVely - c.groundLevel - c.platformPosY) * (c.localscl / oc.localscl))
		case OC_power:
			sys.bcStack.PushI(c.getPower())
		case OC_powermax:
			sys.bcStack.PushI(c.powerMax)
		case OC_playeridexist:
			*sys.bcStack.Top() = sys.playerIDExist(*sys.bcStack.Top())
		case OC_prevstateno:
			sys.bcStack.PushI(c.ss.prevno)
		case OC_projcanceltime:
			*sys.bcStack.Top() = c.projCancelTime(*sys.bcStack.Top())
		case OC_projcontacttime:
			*sys.bcStack.Top() = c.projContactTime(*sys.bcStack.Top())
		case OC_projguardedtime:
			*sys.bcStack.Top() = c.projGuardedTime(*sys.bcStack.Top())
		case OC_projhittime:
			*sys.bcStack.Top() = c.projHitTime(*sys.bcStack.Top())
		case OC_random:
			sys.bcStack.PushI(Rand(0, 999))
		case OC_rightedge:
			sys.bcStack.PushF(c.rightEdge() * (c.localscl / oc.localscl))
		case OC_roundstate:
			sys.bcStack.PushI(sys.roundState())
		case OC_screenheight:
			sys.bcStack.PushF(c.screenHeight())
		case OC_screenpos_x:
			sys.bcStack.PushF((c.screenPosX()) / oc.localscl)
		case OC_screenpos_y:
			sys.bcStack.PushF((c.screenPosY()) / oc.localscl)
		case OC_screenwidth:
			sys.bcStack.PushF(c.screenWidth())
		case OC_selfanimexist:
			*sys.bcStack.Top() = c.selfAnimExist(*sys.bcStack.Top())
		case OC_stateno:
			sys.bcStack.PushI(c.ss.no)
		case OC_statetype:
			sys.bcStack.PushB(c.ss.stateType == StateType(be[i]))
			i++
		case OC_teammode:
			if c.teamside == -1 {
				sys.bcStack.PushB(TM_Single == TeamMode(be[i]))
			} else {
				sys.bcStack.PushB(sys.tmode[c.playerNo&1] == TeamMode(be[i]))
			}
			i++
		case OC_teamside:
			sys.bcStack.PushI(int32(c.teamside) + 1)
		case OC_time:
			sys.bcStack.PushI(c.time())
		case OC_topedge:
			sys.bcStack.PushF(c.topEdge() * (c.localscl / oc.localscl))
		case OC_uniqhitcount:
			sys.bcStack.PushI(c.uniqHitCount)
		case OC_vel_x:
			sys.bcStack.PushF(c.vel[0] * (c.localscl / oc.localscl))
		case OC_vel_y:
			sys.bcStack.PushF(c.vel[1] * (c.localscl / oc.localscl))
		case OC_st_:
			be.run_st(c, &i)
		case OC_const_:
			be.run_const(c, &i, oc)
		case OC_ex_:
			be.run_ex(c, &i, oc)
		case OC_ex2_:
			be.run_ex2(c, &i, oc)
		case OC_var:
			*sys.bcStack.Top() = c.varGet(sys.bcStack.Top().ToI())
		case OC_sysvar:
			*sys.bcStack.Top() = c.sysVarGet(sys.bcStack.Top().ToI())
		case OC_fvar:
			*sys.bcStack.Top() = c.fvarGet(sys.bcStack.Top().ToI())
		case OC_sysfvar:
			*sys.bcStack.Top() = c.sysFvarGet(sys.bcStack.Top().ToI())
		case OC_localvar:
			sys.bcStack.Push(sys.bcVar[uint8(be[i])])
			i++
		}
		c = oc
	}
	return sys.bcStack.Pop()
}
func (be BytecodeExp) run_st(c *Char, i *int) {
	(*i)++
	switch be[*i-1] {
	case OC_st_var:
		v := sys.bcStack.Pop().ToI()
		*sys.bcStack.Top() = c.varSet(sys.bcStack.Top().ToI(), v)
	case OC_st_sysvar:
		v := sys.bcStack.Pop().ToI()
		*sys.bcStack.Top() = c.sysVarSet(sys.bcStack.Top().ToI(), v)
	case OC_st_fvar:
		v := sys.bcStack.Pop().ToF()
		*sys.bcStack.Top() = c.fvarSet(sys.bcStack.Top().ToI(), v)
	case OC_st_sysfvar:
		v := sys.bcStack.Pop().ToF()
		*sys.bcStack.Top() = c.sysFvarSet(sys.bcStack.Top().ToI(), v)
	case OC_st_varadd:
		v := sys.bcStack.Pop().ToI()
		*sys.bcStack.Top() = c.varAdd(sys.bcStack.Top().ToI(), v)
	case OC_st_sysvaradd:
		v := sys.bcStack.Pop().ToI()
		*sys.bcStack.Top() = c.sysVarAdd(sys.bcStack.Top().ToI(), v)
	case OC_st_fvaradd:
		v := sys.bcStack.Pop().ToF()
		*sys.bcStack.Top() = c.fvarAdd(sys.bcStack.Top().ToI(), v)
	case OC_st_sysfvaradd:
		v := sys.bcStack.Pop().ToF()
		*sys.bcStack.Top() = c.sysFvarAdd(sys.bcStack.Top().ToI(), v)
	case OC_st_map:
		v := sys.bcStack.Pop().ToF()
		sys.bcStack.Push(c.mapSet(sys.stringPool[sys.workingState.playerNo].List[*(*int32)(unsafe.Pointer(&be[*i]))], v, 0))
		*i += 4
	}
}
func (be BytecodeExp) run_const(c *Char, i *int, oc *Char) {
	(*i)++
	switch be[*i-1] {
	case OC_const_data_life:
		sys.bcStack.PushI(c.gi().data.life)
	case OC_const_data_power:
		sys.bcStack.PushI(c.gi().data.power)
	case OC_const_data_dizzypoints:
		sys.bcStack.PushI(c.gi().data.dizzypoints)
	case OC_const_data_guardpoints:
		sys.bcStack.PushI(c.gi().data.guardpoints)
	case OC_const_data_attack:
		sys.bcStack.PushI(c.gi().data.attack)
	case OC_const_data_defence:
		sys.bcStack.PushI(c.gi().data.defence)
	case OC_const_data_fall_defence_up:
		sys.bcStack.PushI(c.gi().data.fall.defence_up)
	case OC_const_data_fall_defence_mul:
		sys.bcStack.PushF(1.0 / c.gi().data.fall.defence_mul)
	case OC_const_data_liedown_time:
		sys.bcStack.PushI(c.gi().data.liedown.time)
	case OC_const_data_airjuggle:
		sys.bcStack.PushI(c.gi().data.airjuggle)
	case OC_const_data_sparkno:
		sys.bcStack.PushI(c.gi().data.sparkno)
	case OC_const_data_guard_sparkno:
		sys.bcStack.PushI(c.gi().data.guard.sparkno)
	case OC_const_data_hitsound_channel:
		sys.bcStack.PushI(c.gi().data.hitsound_channel)
	case OC_const_data_guardsound_channel:
		sys.bcStack.PushI(c.gi().data.guardsound_channel)
	case OC_const_data_ko_echo:
		sys.bcStack.PushI(c.gi().data.ko.echo)
	case OC_const_data_intpersistindex:
		sys.bcStack.PushI(c.gi().data.intpersistindex)
	case OC_const_data_floatpersistindex:
		sys.bcStack.PushI(c.gi().data.floatpersistindex)
	case OC_const_size_xscale:
		sys.bcStack.PushF(c.size.xscale)
	case OC_const_size_yscale:
		sys.bcStack.PushF(c.size.yscale)
	case OC_const_size_ground_back:
		sys.bcStack.PushF(c.size.ground.back * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_ground_front:
		sys.bcStack.PushF(c.size.ground.front * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_air_back:
		sys.bcStack.PushF(c.size.air.back * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_air_front:
		sys.bcStack.PushF(c.size.air.front * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_height_stand:
		sys.bcStack.PushF(c.size.height.stand * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_height_crouch:
		sys.bcStack.PushF(c.size.height.crouch * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_height_air_top:
		sys.bcStack.PushF(c.size.height.air[0] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_height_air_bottom:
		sys.bcStack.PushF(c.size.height.air[1] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_height_down:
		sys.bcStack.PushF(c.size.height.down * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_attack_dist_front:
		sys.bcStack.PushF(c.size.attack.dist.front * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_attack_dist_back:
		sys.bcStack.PushF(c.size.attack.dist.back * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_attack_z_width_back:
		sys.bcStack.PushF(c.size.attack.z.width[1] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_attack_z_width_front:
		sys.bcStack.PushF(c.size.attack.z.width[0] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_proj_attack_dist_front:
		sys.bcStack.PushF(c.size.proj.attack.dist.front * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_proj_attack_dist_back:
		sys.bcStack.PushF(c.size.proj.attack.dist.back * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_proj_doscale:
		sys.bcStack.PushI(c.size.proj.doscale)
	case OC_const_size_head_pos_x:
		sys.bcStack.PushF(c.size.head.pos[0] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_head_pos_y:
		sys.bcStack.PushF(c.size.head.pos[1] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_mid_pos_x:
		sys.bcStack.PushF(c.size.mid.pos[0] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_mid_pos_y:
		sys.bcStack.PushF(c.size.mid.pos[1] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_shadowoffset:
		sys.bcStack.PushF(c.size.shadowoffset * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_draw_offset_x:
		sys.bcStack.PushF(c.size.draw.offset[0] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_draw_offset_y:
		sys.bcStack.PushF(c.size.draw.offset[1] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_z_width:
		sys.bcStack.PushF(c.size.z.width * ((320 / c.localcoord) / oc.localscl))
	case OC_const_size_z_enable:
		sys.bcStack.PushB(c.size.z.enable)
	case OC_const_velocity_walk_fwd_x:
		sys.bcStack.PushF(c.gi().velocity.walk.fwd * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_walk_back_x:
		sys.bcStack.PushF(c.gi().velocity.walk.back * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_walk_up_x:
		sys.bcStack.PushF(c.gi().velocity.walk.up.x * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_walk_down_x:
		sys.bcStack.PushF(c.gi().velocity.walk.down.x * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_run_fwd_x:
		sys.bcStack.PushF(c.gi().velocity.run.fwd[0] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_run_fwd_y:
		sys.bcStack.PushF(c.gi().velocity.run.fwd[1] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_run_back_x:
		sys.bcStack.PushF(c.gi().velocity.run.back[0] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_run_back_y:
		sys.bcStack.PushF(c.gi().velocity.run.back[1] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_run_up_x:
		sys.bcStack.PushF(c.gi().velocity.run.up.x * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_run_up_y:
		sys.bcStack.PushF(c.gi().velocity.run.up.y * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_run_down_x:
		sys.bcStack.PushF(c.gi().velocity.run.down.x * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_run_down_y:
		sys.bcStack.PushF(c.gi().velocity.run.down.y * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_jump_y:
		sys.bcStack.PushF(c.gi().velocity.jump.neu[1] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_jump_neu_x:
		sys.bcStack.PushF(c.gi().velocity.jump.neu[0] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_jump_back_x:
		sys.bcStack.PushF(c.gi().velocity.jump.back * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_jump_fwd_x:
		sys.bcStack.PushF(c.gi().velocity.jump.fwd * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_jump_up_x:
		sys.bcStack.PushF(c.gi().velocity.jump.up.x * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_jump_down_x:
		sys.bcStack.PushF(c.gi().velocity.jump.down.x * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_runjump_back_x:
		sys.bcStack.PushF(c.gi().velocity.runjump.back[0] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_runjump_back_y:
		sys.bcStack.PushF(c.gi().velocity.runjump.back[1] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_runjump_y:
		sys.bcStack.PushF(c.gi().velocity.runjump.fwd[1] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_runjump_fwd_x:
		sys.bcStack.PushF(c.gi().velocity.runjump.fwd[0] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_runjump_up_x:
		sys.bcStack.PushF(c.gi().velocity.runjump.up.x * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_runjump_down_x:
		sys.bcStack.PushF(c.gi().velocity.runjump.down.x * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_airjump_y:
		sys.bcStack.PushF(c.gi().velocity.airjump.neu[1] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_airjump_neu_x:
		sys.bcStack.PushF(c.gi().velocity.airjump.neu[0] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_airjump_back_x:
		sys.bcStack.PushF(c.gi().velocity.airjump.back * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_airjump_fwd_x:
		sys.bcStack.PushF(c.gi().velocity.airjump.fwd * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_airjump_up_x:
		sys.bcStack.PushF(c.gi().velocity.airjump.up.x * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_airjump_down_x:
		sys.bcStack.PushF(c.gi().velocity.airjump.down.x * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_air_gethit_groundrecover_x:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.groundrecover[0] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_air_gethit_groundrecover_y:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.groundrecover[1] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_air_gethit_airrecover_mul_x:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.airrecover.mul[0])
	case OC_const_velocity_air_gethit_airrecover_mul_y:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.airrecover.mul[1])
	case OC_const_velocity_air_gethit_airrecover_add_x:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.airrecover.add[0] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_air_gethit_airrecover_add_y:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.airrecover.add[1] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_air_gethit_airrecover_back:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.airrecover.back * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_air_gethit_airrecover_fwd:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.airrecover.fwd * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_air_gethit_airrecover_up:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.airrecover.up * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_air_gethit_airrecover_down:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.airrecover.down * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_air_gethit_ko_add_x:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.ko.add[0] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_air_gethit_ko_add_y:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.ko.add[1] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_air_gethit_ko_ymin:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.ko.ymin * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_ground_gethit_ko_xmul:
		sys.bcStack.PushF(c.gi().velocity.ground.gethit.ko.xmul)
	case OC_const_velocity_ground_gethit_ko_add_x:
		sys.bcStack.PushF(c.gi().velocity.ground.gethit.ko.add[0] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_ground_gethit_ko_add_y:
		sys.bcStack.PushF(c.gi().velocity.ground.gethit.ko.add[1] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_velocity_ground_gethit_ko_ymin:
		sys.bcStack.PushF(c.gi().velocity.ground.gethit.ko.ymin * ((320 / c.localcoord) / oc.localscl))
	case OC_const_movement_airjump_num:
		sys.bcStack.PushI(c.gi().movement.airjump.num)
	case OC_const_movement_airjump_height:
		sys.bcStack.PushI(int32(float32(c.gi().movement.airjump.height) * ((320 / c.localcoord) / oc.localscl)))
	case OC_const_movement_yaccel:
		sys.bcStack.PushF(c.gi().movement.yaccel * ((320 / c.localcoord) / oc.localscl))
	case OC_const_movement_stand_friction:
		sys.bcStack.PushF(c.gi().movement.stand.friction)
	case OC_const_movement_crouch_friction:
		sys.bcStack.PushF(c.gi().movement.crouch.friction)
	case OC_const_movement_stand_friction_threshold:
		sys.bcStack.PushF(c.gi().movement.stand.friction_threshold * ((320 / c.localcoord) / oc.localscl))
	case OC_const_movement_crouch_friction_threshold:
		sys.bcStack.PushF(c.gi().movement.crouch.friction_threshold * ((320 / c.localcoord) / oc.localscl))
	case OC_const_movement_air_gethit_groundlevel:
		sys.bcStack.PushF(c.gi().movement.air.gethit.groundlevel * ((320 / c.localcoord) / oc.localscl))
	case OC_const_movement_air_gethit_groundrecover_ground_threshold:
		sys.bcStack.PushF(
			c.gi().movement.air.gethit.groundrecover.ground.threshold * ((320 / c.localcoord) / oc.localscl))
	case OC_const_movement_air_gethit_groundrecover_groundlevel:
		sys.bcStack.PushF(c.gi().movement.air.gethit.groundrecover.groundlevel * ((320 / c.localcoord) / oc.localscl))
	case OC_const_movement_air_gethit_airrecover_threshold:
		sys.bcStack.PushF(c.gi().movement.air.gethit.airrecover.threshold * ((320 / c.localcoord) / oc.localscl))
	case OC_const_movement_air_gethit_airrecover_yaccel:
		sys.bcStack.PushF(c.gi().movement.air.gethit.airrecover.yaccel * ((320 / c.localcoord) / oc.localscl))
	case OC_const_movement_air_gethit_trip_groundlevel:
		sys.bcStack.PushF(c.gi().movement.air.gethit.trip.groundlevel * ((320 / c.localcoord) / oc.localscl))
	case OC_const_movement_down_bounce_offset_x:
		sys.bcStack.PushF(c.gi().movement.down.bounce.offset[0] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_movement_down_bounce_offset_y:
		sys.bcStack.PushF(c.gi().movement.down.bounce.offset[1] * ((320 / c.localcoord) / oc.localscl))
	case OC_const_movement_down_bounce_yaccel:
		sys.bcStack.PushF(c.gi().movement.down.bounce.yaccel * ((320 / c.localcoord) / oc.localscl))
	case OC_const_movement_down_bounce_groundlevel:
		sys.bcStack.PushF(c.gi().movement.down.bounce.groundlevel * ((320 / c.localcoord) / oc.localscl))
	case OC_const_movement_down_friction_threshold:
		sys.bcStack.PushF(c.gi().movement.down.friction_threshold * ((320 / c.localcoord) / oc.localscl))
	case OC_const_authorname:
		sys.bcStack.PushB(c.gi().authorLow ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_displayname:
		sys.bcStack.PushB(c.gi().displaynameLow ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_name:
		sys.bcStack.PushB(c.gi().nameLow ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_p2name:
		p2 := c.p2()
		sys.bcStack.PushB(p2 != nil && p2.gi().nameLow ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_p3name:
		p3 := c.partner(0, false)
		sys.bcStack.PushB(p3 != nil && p3.gi().nameLow ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_p4name:
		p4 := sys.charList.enemyNear(c, 1, true, true, false)
		sys.bcStack.PushB(p4 != nil && !(p4.scf(SCF_ko) && p4.scf(SCF_over)) &&
			p4.gi().nameLow ==
				sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
					unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_p5name:
		p5 := c.partner(1, false)
		sys.bcStack.PushB(p5 != nil && p5.gi().nameLow ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_p6name:
		p6 := sys.charList.enemyNear(c, 2, true, true, false)
		sys.bcStack.PushB(p6 != nil && !(p6.scf(SCF_ko) && p6.scf(SCF_over)) &&
			p6.gi().nameLow ==
				sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
					unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_p7name:
		p7 := c.partner(2, false)
		sys.bcStack.PushB(p7 != nil && p7.gi().nameLow ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_p8name:
		p8 := sys.charList.enemyNear(c, 3, true, true, false)
		sys.bcStack.PushB(p8 != nil && !(p8.scf(SCF_ko) && p8.scf(SCF_over)) &&
			p8.gi().nameLow ==
				sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
					unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_stagevar_info_name:
		sys.bcStack.PushB(sys.stage.nameLow ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_stagevar_info_displayname:
		sys.bcStack.PushB(sys.stage.displaynameLow ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_stagevar_info_author:
		sys.bcStack.PushB(sys.stage.authorLow ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_stagevar_camera_boundleft:
		sys.bcStack.PushI(sys.stage.stageCamera.boundleft)
	case OC_const_stagevar_camera_boundright:
		sys.bcStack.PushI(sys.stage.stageCamera.boundright)
	case OC_const_stagevar_camera_boundhigh:
		sys.bcStack.PushI(sys.stage.stageCamera.boundhigh)
	case OC_const_stagevar_camera_boundlow:
		sys.bcStack.PushI(sys.stage.stageCamera.boundlow)
	case OC_const_stagevar_camera_verticalfollow:
		sys.bcStack.PushF(sys.stage.stageCamera.verticalfollow)
	case OC_const_stagevar_camera_floortension:
		sys.bcStack.PushI(sys.stage.stageCamera.floortension)
	case OC_const_stagevar_camera_tensionhigh:
		sys.bcStack.PushI(sys.stage.stageCamera.tensionhigh)
	case OC_const_stagevar_camera_tensionlow:
		sys.bcStack.PushI(sys.stage.stageCamera.tensionlow)
	case OC_const_stagevar_camera_tension:
		sys.bcStack.PushI(sys.stage.stageCamera.tension)
	case OC_const_stagevar_camera_startzoom:
		sys.bcStack.PushF(sys.stage.stageCamera.startzoom)
	case OC_const_stagevar_camera_zoomout:
		sys.bcStack.PushF(sys.stage.stageCamera.zoomout)
	case OC_const_stagevar_camera_zoomin:
		sys.bcStack.PushF(sys.stage.stageCamera.zoomin)
	case OC_const_stagevar_camera_ytension_enable:
		sys.bcStack.PushB(sys.stage.stageCamera.ytensionenable)
	case OC_const_stagevar_playerinfo_leftbound:
		sys.bcStack.PushF(sys.stage.leftbound)
	case OC_const_stagevar_playerinfo_rightbound:
		sys.bcStack.PushF(sys.stage.rightbound)
	case OC_const_stagevar_scaling_topscale:
		sys.bcStack.PushF(sys.stage.stageCamera.ztopscale)
	case OC_const_stagevar_bound_screenleft:
		sys.bcStack.PushI(sys.stage.screenleft)
	case OC_const_stagevar_bound_screenright:
		sys.bcStack.PushI(sys.stage.screenright)
	case OC_const_stagevar_stageinfo_localcoord_x:
		sys.bcStack.PushI(sys.stage.stageCamera.localcoord[0])
	case OC_const_stagevar_stageinfo_localcoord_y:
		sys.bcStack.PushI(sys.stage.stageCamera.localcoord[1])
	case OC_const_stagevar_stageinfo_xscale:
		sys.bcStack.PushF(sys.stage.scale[0])
	case OC_const_stagevar_stageinfo_yscale:
		sys.bcStack.PushF(sys.stage.scale[1])
	case OC_const_stagevar_stageinfo_zoffset:
		sys.bcStack.PushI(sys.stage.stageCamera.zoffset)
	case OC_const_stagevar_stageinfo_zoffsetlink:
		sys.bcStack.PushI(sys.stage.zoffsetlink)
	case OC_const_stagevar_shadow_intensity:
		sys.bcStack.PushI(sys.stage.sdw.intensity)
	case OC_const_stagevar_shadow_color_r:
		sys.bcStack.PushI(int32((sys.stage.sdw.color & 0xFF0000) >> 16))
	case OC_const_stagevar_shadow_color_g:
		sys.bcStack.PushI(int32((sys.stage.sdw.color & 0xFF00) >> 8))
	case OC_const_stagevar_shadow_color_b:
		sys.bcStack.PushI(int32(sys.stage.sdw.color & 0xFF))
	case OC_const_stagevar_shadow_yscale:
		sys.bcStack.PushF(sys.stage.sdw.yscale)
	case OC_const_stagevar_shadow_fade_range_begin:
		sys.bcStack.PushI(sys.stage.sdw.fadebgn)
	case OC_const_stagevar_shadow_fade_range_end:
		sys.bcStack.PushI(sys.stage.sdw.fadeend)
	case OC_const_stagevar_shadow_xshear:
		sys.bcStack.PushF(sys.stage.sdw.xshear)
	case OC_const_stagevar_reflection_intensity:
		sys.bcStack.PushI(sys.stage.reflection)
	case OC_const_constants:
		sys.bcStack.PushF(c.gi().constants[sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
			unsafe.Pointer(&be[*i]))]])
		*i += 4
	case OC_const_stage_constants:
		sys.bcStack.PushF(sys.stage.constants[sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
			unsafe.Pointer(&be[*i]))]])
		*i += 4
	default:
		sys.errLog.Printf("%v\n", be[*i-1])
		c.panic()
	}
}
func (be BytecodeExp) run_ex(c *Char, i *int, oc *Char) {
	(*i)++
	switch be[*i-1] {
	case OC_ex_p2dist_x:
		sys.bcStack.Push(c.rdDistX(c.p2(), oc))
	case OC_ex_p2dist_y:
		sys.bcStack.Push(c.rdDistY(c.p2(), oc))
	case OC_ex_p2bodydist_x:
		sys.bcStack.Push(c.p2BodyDistX(oc))
	case OC_ex_p2bodydist_y:
		sys.bcStack.Push(c.p2BodyDistY(oc))
	case OC_ex_parentdist_x:
		sys.bcStack.Push(c.rdDistX(c.parent(), oc))
	case OC_ex_parentdist_y:
		sys.bcStack.Push(c.rdDistY(c.parent(), oc))
	case OC_ex_rootdist_x:
		sys.bcStack.Push(c.rdDistX(c.root(), oc))
	case OC_ex_rootdist_y:
		sys.bcStack.Push(c.rdDistY(c.root(), oc))
	case OC_ex_win:
		sys.bcStack.PushB(c.win())
	case OC_ex_winko:
		sys.bcStack.PushB(c.winKO())
	case OC_ex_wintime:
		sys.bcStack.PushB(c.winTime())
	case OC_ex_winperfect:
		sys.bcStack.PushB(c.winPerfect())
	case OC_ex_winspecial:
		sys.bcStack.PushB(c.winType(WT_Special))
	case OC_ex_winhyper:
		sys.bcStack.PushB(c.winType(WT_Hyper))
	case OC_ex_lose:
		sys.bcStack.PushB(c.lose())
	case OC_ex_loseko:
		sys.bcStack.PushB(c.loseKO())
	case OC_ex_losetime:
		sys.bcStack.PushB(c.loseTime())
	case OC_ex_drawgame:
		sys.bcStack.PushB(c.drawgame())
	case OC_ex_matchover:
		sys.bcStack.PushB(sys.matchOver())
	case OC_ex_matchno:
		sys.bcStack.PushI(sys.match)
	case OC_ex_roundno:
		sys.bcStack.PushI(sys.round)
	case OC_ex_roundsexisted:
		sys.bcStack.PushI(c.roundsExisted())
	case OC_ex_ishometeam:
		sys.bcStack.PushB(c.teamside == sys.home)
	case OC_ex_tickspersecond:
		sys.bcStack.PushI(int32(float32(FPS) * sys.gameSpeed * sys.accel))
	case OC_ex_const240p:
		*sys.bcStack.Top() = c.constp(320, sys.bcStack.Top().ToF())
	case OC_ex_const480p:
		*sys.bcStack.Top() = c.constp(640, sys.bcStack.Top().ToF())
	case OC_ex_const720p:
		*sys.bcStack.Top() = c.constp(1280, sys.bcStack.Top().ToF())
	case OC_ex_const1080p:
		*sys.bcStack.Top() = c.constp(1920, sys.bcStack.Top().ToF())
	case OC_ex_gethitvar_animtype:
		sys.bcStack.PushI(int32(c.ghv.animtype))
	case OC_ex_gethitvar_air_animtype:
		sys.bcStack.PushI(int32(c.ghv.airanimtype))
	case OC_ex_gethitvar_ground_animtype:
		sys.bcStack.PushI(int32(c.ghv.groundanimtype))
	case OC_ex_gethitvar_fall_animtype:
		sys.bcStack.PushI(int32(c.ghv.fall.animtype))
	case OC_ex_gethitvar_type:
		sys.bcStack.PushI(int32(c.ghv._type))
	case OC_ex_gethitvar_airtype:
		sys.bcStack.PushI(int32(c.ghv.airtype))
	case OC_ex_gethitvar_groundtype:
		sys.bcStack.PushI(int32(c.ghv.groundtype))
	case OC_ex_gethitvar_damage:
		sys.bcStack.PushI(c.ghv.damage)
	case OC_ex_gethitvar_guardcount:
		sys.bcStack.PushI(c.ghv.guardcount)
	case OC_ex_gethitvar_hitcount:
		sys.bcStack.PushI(c.ghv.hitcount)
	case OC_ex_gethitvar_fallcount:
		sys.bcStack.PushI(c.ghv.fallcount)
	case OC_ex_gethitvar_hitshaketime:
		sys.bcStack.PushI(c.ghv.hitshaketime)
	case OC_ex_gethitvar_hittime:
		sys.bcStack.PushI(c.ghv.hittime)
	case OC_ex_gethitvar_slidetime:
		sys.bcStack.PushI(c.ghv.slidetime)
	case OC_ex_gethitvar_ctrltime:
		sys.bcStack.PushI(c.ghv.ctrltime)
	case OC_ex_gethitvar_recovertime:
		sys.bcStack.PushI(c.recoverTime)
	case OC_ex_gethitvar_xoff:
		sys.bcStack.PushF(c.ghv.xoff * (c.localscl / oc.localscl))
	case OC_ex_gethitvar_yoff:
		sys.bcStack.PushF(c.ghv.yoff * (c.localscl / oc.localscl))
	case OC_ex_gethitvar_xvel:
		sys.bcStack.PushF(c.ghv.xvel * c.facing * (c.localscl / oc.localscl))
	case OC_ex_gethitvar_yvel:
		sys.bcStack.PushF(c.ghv.yvel * (c.localscl / oc.localscl))
	case OC_ex_gethitvar_yaccel:
		sys.bcStack.PushF(c.ghv.getYaccel(oc) * (c.localscl / oc.localscl))
	case OC_ex_gethitvar_chainid:
		sys.bcStack.PushI(c.ghv.chainId())
	case OC_ex_gethitvar_guarded:
		sys.bcStack.PushB(c.ghv.guarded)
	case OC_ex_gethitvar_isbound:
		sys.bcStack.PushB(c.isBound())
	case OC_ex_gethitvar_fall:
		sys.bcStack.PushB(c.ghv.fallf)
	case OC_ex_gethitvar_fall_damage:
		sys.bcStack.PushI(c.ghv.fall.damage)
	case OC_ex_gethitvar_fall_xvel:
		sys.bcStack.PushF(c.ghv.fall.xvel() * (c.localscl / oc.localscl))
	case OC_ex_gethitvar_fall_yvel:
		sys.bcStack.PushF(c.ghv.fall.yvelocity * (c.localscl / oc.localscl))
	case OC_ex_gethitvar_fall_recover:
		sys.bcStack.PushB(c.ghv.fall.recover)
	case OC_ex_gethitvar_fall_time:
		sys.bcStack.PushI(c.fallTime)
	case OC_ex_gethitvar_fall_recovertime:
		sys.bcStack.PushI(c.ghv.fall.recovertime)
	case OC_ex_gethitvar_fall_kill:
		sys.bcStack.PushB(c.ghv.fall.kill)
	case OC_ex_gethitvar_fall_envshake_time:
		sys.bcStack.PushI(c.ghv.fall.envshake_time)
	case OC_ex_gethitvar_fall_envshake_freq:
		sys.bcStack.PushF(c.ghv.fall.envshake_freq)
	case OC_ex_gethitvar_fall_envshake_ampl:
		sys.bcStack.PushI(int32(float32(c.ghv.fall.envshake_ampl) * (c.localscl / oc.localscl)))
	case OC_ex_gethitvar_fall_envshake_phase:
		sys.bcStack.PushF(c.ghv.fall.envshake_phase * (c.localscl / oc.localscl))
	case OC_ex_gethitvar_fall_envshake_mul:
		sys.bcStack.PushF(c.ghv.fall.envshake_mul)
	case OC_ex_gethitvar_attr:
		sys.bcStack.PushI(c.ghv.attr)
	case OC_ex_gethitvar_dizzypoints:
		sys.bcStack.PushI(c.ghv.dizzypoints)
	case OC_ex_gethitvar_guardpoints:
		sys.bcStack.PushI(c.ghv.guardpoints)
	case OC_ex_gethitvar_id:
		sys.bcStack.PushI(c.ghv.id)
	case OC_ex_gethitvar_playerno:
		sys.bcStack.PushI(int32(c.ghv.playerNo) + 1)
	case OC_ex_gethitvar_redlife:
		sys.bcStack.PushI(c.ghv.redlife)
	case OC_ex_gethitvar_score:
		sys.bcStack.PushF(c.ghv.score)
	case OC_ex_gethitvar_hitdamage:
		sys.bcStack.PushI(c.ghv.hitdamage)
	case OC_ex_gethitvar_guarddamage:
		sys.bcStack.PushI(c.ghv.guarddamage)
	case OC_ex_gethitvar_power:
		sys.bcStack.PushI(c.ghv.power)
	case OC_ex_gethitvar_hitpower:
		sys.bcStack.PushI(c.ghv.hitpower)
	case OC_ex_gethitvar_guardpower:
		sys.bcStack.PushI(c.ghv.guardpower)
	case OC_ex_gethitvar_kill:
		sys.bcStack.PushB(c.ghv.kill)
	case OC_ex_gethitvar_priority:
		sys.bcStack.PushI(c.ghv.priority)
	case OC_ex_gethitvar_facing:
		sys.bcStack.PushI(c.ghv.facing)
	case OC_ex_gethitvar_ground_velocity_x:
		sys.bcStack.PushF(c.ghv.ground_velocity[0] * c.facing * (c.localscl / oc.localscl))
	case OC_ex_gethitvar_ground_velocity_y:
		sys.bcStack.PushF(c.ghv.ground_velocity[1] * (c.localscl / oc.localscl))
	case OC_ex_gethitvar_air_velocity_x:
		sys.bcStack.PushF(c.ghv.air_velocity[0] * c.facing * (c.localscl / oc.localscl))
	case OC_ex_gethitvar_air_velocity_y:
		sys.bcStack.PushF(c.ghv.air_velocity[1] * (c.localscl / oc.localscl))
	case OC_ex_gethitvar_down_velocity_x:
		sys.bcStack.PushF(c.ghv.down_velocity[0] * c.facing * (c.localscl / oc.localscl))
	case OC_ex_gethitvar_down_velocity_y:
		sys.bcStack.PushF(c.ghv.down_velocity[1] * (c.localscl / oc.localscl))
	case OC_ex_gethitvar_guard_velocity_x:
		sys.bcStack.PushF(c.ghv.guard_velocity * c.facing * (c.localscl / oc.localscl))
	case OC_ex_gethitvar_airguard_velocity_x:
		sys.bcStack.PushF(c.ghv.airguard_velocity[0] * c.facing * (c.localscl / oc.localscl))
	case OC_ex_gethitvar_airguard_velocity_y:
		sys.bcStack.PushF(c.ghv.airguard_velocity[1] * (c.localscl / oc.localscl))
	case OC_ex_gethitvar_frame:
		sys.bcStack.PushB(c.ghv.frame)
	case OC_ex_ailevelf:
		if !c.asf(ASF_noailevel) {
			sys.bcStack.PushF(c.aiLevel())
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_animelemlength:
		if f := c.anim.CurrentFrame(); f != nil {
			sys.bcStack.PushI(f.Time)
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_animframe_alphadest:
		if f := c.anim.CurrentFrame(); f != nil {
			sys.bcStack.PushI(int32(f.DstAlpha))
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_animframe_angle:
		if f := c.anim.CurrentFrame(); f != nil {
			if len(f.Ex) > 2 && len(f.Ex[2]) > 2 { // Anim.go code could be refactored so these are easier to read
				sys.bcStack.PushF(f.Ex[2][2])
			} else {
				sys.bcStack.PushF(0)
			}
		}
	case OC_ex_animframe_alphasource:
		if f := c.anim.CurrentFrame(); f != nil {
			sys.bcStack.PushI(int32(f.SrcAlpha))
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_animframe_group:
		if f := c.anim.CurrentFrame(); f != nil {
			sys.bcStack.PushI(int32(f.Group))
		} else {
			sys.bcStack.PushI(-1)
		}
	case OC_ex_animframe_hflip:
		if f := c.anim.CurrentFrame(); f != nil {
			sys.bcStack.PushB(f.H < 0)
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_animframe_image:
		if f := c.anim.CurrentFrame(); f != nil {
			sys.bcStack.PushI(int32(f.Number))
		} else {
			sys.bcStack.PushI(-1)
		}
	case OC_ex_animframe_time: // Same as AnimElemLength
		if f := c.anim.CurrentFrame(); f != nil {
			sys.bcStack.PushI(f.Time)
		} else {
			sys.bcStack.PushI(-1)
		}
	case OC_ex_animframe_vflip:
		if f := c.anim.CurrentFrame(); f != nil {
			sys.bcStack.PushB(f.V < 0)
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_animframe_xoffset:
		if f := c.anim.CurrentFrame(); f != nil {
			sys.bcStack.PushI(int32(-f.X))
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_animframe_xscale:
		if f := c.anim.CurrentFrame(); f != nil {
			if len(f.Ex) > 2 {
				sys.bcStack.PushF(f.Ex[2][0])
			} else {
				sys.bcStack.PushF(0)
			}
		}
	case OC_ex_animframe_yoffset:
		if f := c.anim.CurrentFrame(); f != nil {
			sys.bcStack.PushI(int32(-f.Y))
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_animframe_yscale:
		if f := c.anim.CurrentFrame(); f != nil {
			if len(f.Ex) > 2 && len(f.Ex[2]) > 1 {
				sys.bcStack.PushF(f.Ex[2][1])
			} else {
				sys.bcStack.PushF(0)
			}
		}
	case OC_ex_animframe_numclsn1:
		if f := c.anim.CurrentFrame(); f != nil {
			sys.bcStack.PushI(int32(len(f.Clsn1()) / 4))
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_animframe_numclsn2:
		if f := c.anim.CurrentFrame(); f != nil {
			sys.bcStack.PushI(int32(len(f.Clsn2()) / 4))
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_animlength:
		sys.bcStack.PushI(c.anim.totaltime)
	case OC_ex_attack:
		sys.bcStack.PushF(c.attackMul * 100)
	case OC_ex_inputtime_B:
		if c.keyctrl[0] && c.cmd != nil {
			sys.bcStack.PushI(c.cmd[0].Buffer.Bb)
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_inputtime_D:
		if c.keyctrl[0] && c.cmd != nil {
			sys.bcStack.PushI(c.cmd[0].Buffer.Db)
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_inputtime_F:
		if c.keyctrl[0] && c.cmd != nil {
			sys.bcStack.PushI(c.cmd[0].Buffer.Fb)
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_inputtime_U:
		if c.keyctrl[0] && c.cmd != nil {
			sys.bcStack.PushI(c.cmd[0].Buffer.Ub)
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_inputtime_L:
		if c.keyctrl[0] && c.cmd != nil {
			sys.bcStack.PushI(c.cmd[0].Buffer.Lb)
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_inputtime_R:
		if c.keyctrl[0] && c.cmd != nil {
			sys.bcStack.PushI(c.cmd[0].Buffer.Rb)
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_inputtime_a:
		if c.keyctrl[0] && c.cmd != nil {
			sys.bcStack.PushI(c.cmd[0].Buffer.ab)
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_inputtime_b:
		if c.keyctrl[0] && c.cmd != nil {
			sys.bcStack.PushI(c.cmd[0].Buffer.bb)
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_inputtime_c:
		if c.keyctrl[0] && c.cmd != nil {
			sys.bcStack.PushI(c.cmd[0].Buffer.cb)
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_inputtime_x:
		if c.keyctrl[0] && c.cmd != nil {
			sys.bcStack.PushI(c.cmd[0].Buffer.xb)
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_inputtime_y:
		if c.keyctrl[0] && c.cmd != nil {
			sys.bcStack.PushI(c.cmd[0].Buffer.yb)
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_inputtime_z:
		if c.keyctrl[0] && c.cmd != nil {
			sys.bcStack.PushI(c.cmd[0].Buffer.zb)
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_inputtime_s:
		if c.keyctrl[0] && c.cmd != nil {
			sys.bcStack.PushI(c.cmd[0].Buffer.sb)
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_inputtime_d:
		if c.keyctrl[0] && c.cmd != nil {
			sys.bcStack.PushI(c.cmd[0].Buffer.db)
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_inputtime_w:
		if c.keyctrl[0] && c.cmd != nil {
			sys.bcStack.PushI(c.cmd[0].Buffer.wb)
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_inputtime_m:
		if c.keyctrl[0] && c.cmd != nil {
			sys.bcStack.PushI(c.cmd[0].Buffer.mb)
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_combocount:
		sys.bcStack.PushI(c.comboCount())
	case OC_ex_consecutivewins:
		sys.bcStack.PushI(c.consecutiveWins())
	case OC_ex_defence:
		sys.bcStack.PushF(float32(c.finalDefense * 100))
	case OC_ex_dizzy:
		sys.bcStack.PushB(c.scf(SCF_dizzy))
	case OC_ex_dizzypoints:
		sys.bcStack.PushI(c.dizzyPoints)
	case OC_ex_dizzypointsmax:
		sys.bcStack.PushI(c.dizzyPointsMax)
	case OC_ex_drawpalno:
		sys.bcStack.PushI(c.gi().drawpalno)
	case OC_ex_fightscreenvar_info_author:
		sys.bcStack.PushB(sys.lifebar.authorLow ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_ex_fightscreenvar_info_name:
		sys.bcStack.PushB(sys.lifebar.nameLow ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_ex_fightscreenvar_round_ctrl_time:
		sys.bcStack.PushI(sys.lifebar.ro.ctrl_time)
	case OC_ex_fightscreenvar_round_over_hittime:
		sys.bcStack.PushI(sys.lifebar.ro.over_hittime)
	case OC_ex_fightscreenvar_round_over_time:
		sys.bcStack.PushI(sys.lifebar.ro.over_time)
	case OC_ex_fightscreenvar_round_over_waittime:
		sys.bcStack.PushI(sys.lifebar.ro.over_waittime)
	case OC_ex_fightscreenvar_round_over_wintime:
		sys.bcStack.PushI(sys.lifebar.ro.over_wintime)
	case OC_ex_fightscreenvar_round_slow_time:
		sys.bcStack.PushI(sys.lifebar.ro.slow_time)
	case OC_ex_fightscreenvar_round_start_waittime:
		sys.bcStack.PushI(sys.lifebar.ro.start_waittime)
	case OC_ex_fighttime:
		sys.bcStack.PushI(sys.gameTime)
	case OC_ex_firstattack:
		sys.bcStack.PushB(sys.firstAttack[c.teamside] == c.playerNo)
	case OC_ex_framespercount:
		sys.bcStack.PushI(sys.lifebar.ti.framespercount)
	case OC_ex_float:
		*sys.bcStack.Top() = BytecodeFloat(sys.bcStack.Top().ToF())
	case OC_ex_gamefps:
		sys.bcStack.PushF(sys.gameFPS)
	case OC_ex_gamemode:
		sys.bcStack.PushB(strings.ToLower(sys.gameMode) ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_ex_getplayerid:
		sys.bcStack.Top().SetI(c.getPlayerID(int(sys.bcStack.Top().ToI())))
	case OC_ex_groundangle:
		sys.bcStack.PushF(c.groundAngle)
	case OC_ex_guardbreak:
		sys.bcStack.PushB(c.scf(SCF_guardbreak))
	case OC_ex_guardpoints:
		sys.bcStack.PushI(c.guardPoints)
	case OC_ex_guardpointsmax:
		sys.bcStack.PushI(c.guardPointsMax)
	case OC_ex_helperid:
		sys.bcStack.PushI(c.helperId)
	case OC_ex_helpername:
		sys.bcStack.PushB(c.helperIndex != 0 && strings.ToLower(c.name) ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_ex_helperindexexist:
		*sys.bcStack.Top() = c.helperByIndexExist(*sys.bcStack.Top())
	case OC_ex_hitoverridden:
		sys.bcStack.PushB(c.hoIdx >= 0)
	case OC_ex_incustomstate:
		sys.bcStack.PushB(c.ss.sb.playerNo != c.playerNo)
	case OC_ex_indialogue:
		sys.bcStack.PushB(sys.dialogueFlg)
	case OC_ex_isassertedchar:
		sys.bcStack.PushB(c.asf(AssertSpecialFlag((*(*int64)(unsafe.Pointer(&be[*i]))))))
		*i += 8
	case OC_ex_isassertedglobal:
		sys.bcStack.PushB(sys.gsf(GlobalSpecialFlag((*(*int32)(unsafe.Pointer(&be[*i]))))))
		*i += 4
	case OC_ex_ishost:
		sys.bcStack.PushB(c.isHost())
	case OC_ex_jugglepoints:
		*sys.bcStack.Top() = c.jugglePoints(*sys.bcStack.Top())
	case OC_ex_localcoord_x:
		sys.bcStack.PushF(sys.cgi[c.playerNo].localcoord[0])
	case OC_ex_localcoord_y:
		sys.bcStack.PushF(sys.cgi[c.playerNo].localcoord[1])
	case OC_ex_localscale:
		sys.bcStack.PushF(c.localscl)
	case OC_ex_majorversion:
		sys.bcStack.PushB(c.gi().mugenver[0] == 1)
	case OC_ex_maparray:
		sys.bcStack.PushF(c.mapArray[sys.stringPool[sys.workingState.playerNo].List[*(*int32)(unsafe.Pointer(&be[*i]))]])
		*i += 4
	case OC_ex_max:
		v2 := sys.bcStack.Pop()
		be.max(sys.bcStack.Top(), v2)
	case OC_ex_min:
		v2 := sys.bcStack.Pop()
		be.min(sys.bcStack.Top(), v2)
	case OC_ex_numplayer:
		sys.bcStack.PushI(c.numPlayer())
	case OC_ex_clamp:
		v3 := sys.bcStack.Pop()
		v2 := sys.bcStack.Pop()
		be.clamp(sys.bcStack.Top(), v2, v3)
	case OC_ex_atan2:
		v2 := sys.bcStack.Pop()
		be.atan2(sys.bcStack.Top(), v2)
	case OC_ex_sign:
		be.sign(sys.bcStack.Top())
	case OC_ex_rad:
		be.rad(sys.bcStack.Top())
	case OC_ex_deg:
		be.deg(sys.bcStack.Top())
	case OC_ex_lastplayerid:
		sys.bcStack.PushI(sys.nextCharId - 1)
	case OC_ex_lerp:
		v3 := sys.bcStack.Pop()
		v2 := sys.bcStack.Pop()
		be.lerp(sys.bcStack.Top(), v2, v3)
	case OC_ex_memberno:
		sys.bcStack.PushI(int32(c.memberNo) + 1)
	case OC_ex_movecountered:
		sys.bcStack.PushI(c.moveCountered())
	case OC_ex_mugenversion:
		sys.bcStack.PushF(c.mugenVersion())
	case OC_ex_pausetime:
		sys.bcStack.PushI(c.pauseTime())
	case OC_ex_physics:
		sys.bcStack.PushB(c.ss.physics == StateType(be[*i]))
		*i++
	case OC_ex_playerno:
		sys.bcStack.PushI(int32(c.playerNo) + 1)
	case OC_ex_playerindexexist:
		*sys.bcStack.Top() = sys.playerIndexExist(*sys.bcStack.Top())
	case OC_ex_randomrange:
		v2 := sys.bcStack.Pop()
		be.random(sys.bcStack.Top(), v2)
	case OC_ex_ratiolevel:
		sys.bcStack.PushI(c.ocd().ratioLevel)
	case OC_ex_receiveddamage:
		sys.bcStack.PushI(c.receivedDmg)
	case OC_ex_receivedhits:
		sys.bcStack.PushI(c.receivedHits)
	case OC_ex_redlife:
		sys.bcStack.PushI(c.redLife)
	case OC_ex_round:
		v2 := sys.bcStack.Pop()
		be.round(sys.bcStack.Top(), v2)
	case OC_ex_roundtype:
		sys.bcStack.PushI(c.roundType())
	case OC_ex_score:
		sys.bcStack.PushF(c.score())
	case OC_ex_scoretotal:
		sys.bcStack.PushF(c.scoreTotal())
	case OC_ex_selfstatenoexist:
		*sys.bcStack.Top() = c.selfStatenoExist(*sys.bcStack.Top())
	case OC_ex_sprpriority:
		sys.bcStack.PushI(c.sprPriority)
	case OC_ex_stagebackedgedist:
		sys.bcStack.PushF(c.stageBackEdgeDist() * (c.localscl / oc.localscl))
	case OC_ex_stagefrontedgedist:
		sys.bcStack.PushF(c.stageFrontEdgeDist() * (c.localscl / oc.localscl))
	case OC_ex_stagetime:
		sys.bcStack.PushI(sys.stage.stageTime)
	case OC_ex_standby:
		sys.bcStack.PushB(c.scf(SCF_standby))
	case OC_ex_teamleader:
		sys.bcStack.PushI(int32(c.teamLeader()))
	case OC_ex_teamsize:
		sys.bcStack.PushI(c.teamSize())
	case OC_ex_timeelapsed:
		sys.bcStack.PushI(timeElapsed())
	case OC_ex_timeremaining:
		sys.bcStack.PushI(timeRemaining())
	case OC_ex_timetotal:
		sys.bcStack.PushI(timeTotal())
	case OC_ex_playercount:
		sys.bcStack.PushI(sys.playercount())
	case OC_ex_pos_z:
		sys.bcStack.PushF(c.pos[2] * (c.localscl / oc.localscl))
	case OC_ex_vel_z:
		sys.bcStack.PushF(c.vel[2] * (c.localscl / oc.localscl))
	case OC_ex_prevanim:
		sys.bcStack.PushI(c.prevAnimNo)
	case OC_ex_prevmovetype:
		sys.bcStack.PushB(c.ss.prevMoveType == MoveType(be[*i])<<15)
		*i++
	case OC_ex_prevstatetype:
		sys.bcStack.PushB(c.ss.prevStateType == StateType(be[*i]))
		*i++
	case OC_ex_reversaldefattr:
		sys.bcStack.PushB(c.reversalDefAttr(*(*int32)(unsafe.Pointer(&be[*i]))))
		*i += 4
	case OC_ex_bgmlength:
		if sys.bgm.streamer == nil {
			sys.bcStack.PushI(0)
		} else {
			sys.bcStack.PushI(int32(sys.bgm.streamer.Len()))
		}
	case OC_ex_bgmposition:
		if sys.bgm.streamer == nil {
			sys.bcStack.PushI(0)
		} else {
			sys.bcStack.PushI(int32(sys.bgm.streamer.Position()))
		}
	case OC_ex_airjumpcount:
		sys.bcStack.PushI(c.airJumpCount)
	case OC_ex_envshakevar_time:
		sys.bcStack.PushI(sys.envShake.time)
	case OC_ex_envshakevar_freq:
		sys.bcStack.PushF(sys.envShake.freq / float32(math.Pi) * 180)
	case OC_ex_envshakevar_ampl:
		sys.bcStack.PushF(float32(math.Abs(float64(sys.envShake.ampl / oc.localscl))))
	case OC_ex_angle:
		sys.bcStack.PushF(c.angleTrg)
	case OC_ex_scale_x:
		sys.bcStack.PushF(c.angleScaleTrg[0])
	case OC_ex_scale_y:
		sys.bcStack.PushF(c.angleScaleTrg[1])
	case OC_ex_offset_x:
		sys.bcStack.PushF(c.offsetTrg[0])
	case OC_ex_offset_y:
		sys.bcStack.PushF(c.offsetTrg[1])
	case OC_ex_alpha_s:
		sys.bcStack.PushI(c.alphaTrg[0])
	case OC_ex_alpha_d:
		sys.bcStack.PushI(c.alphaTrg[1])
	case OC_ex_selfcommand:
		if c.cmd == nil {
			sys.bcStack.PushB(false)
		} else {
			cmd, ok := c.cmd[sys.workingState.playerNo].Names[sys.stringPool[sys.workingState.playerNo].List[*(*int32)(unsafe.Pointer(&be[*i]))]]
			ok = ok && c.command(sys.workingState.playerNo, cmd)
			sys.bcStack.PushB(ok)
		}
		*i += 4
	case OC_ex_guardcount:
		sys.bcStack.PushI(c.guardCount)
	case OC_ex_movehitvar_frame:
		sys.bcStack.PushB(c.mhv.frame)
	case OC_ex_movehitvar_cornerpush:
		sys.bcStack.PushF(c.mhv.cornerpush)
	case OC_ex_movehitvar_id:
		sys.bcStack.PushI(c.mhv.id)
	case OC_ex_movehitvar_overridden:
		sys.bcStack.PushB(c.mhv.overridden)
	case OC_ex_movehitvar_playerno:
		sys.bcStack.PushI(int32(c.mhv.playerNo))
	case OC_ex_movehitvar_spark_x:
		sys.bcStack.PushF(c.mhv.sparkxy[0] * (c.localscl / oc.localscl))
	case OC_ex_movehitvar_spark_y:
		sys.bcStack.PushF(c.mhv.sparkxy[1] * (c.localscl / oc.localscl))
	case OC_ex_movehitvar_uniqhit:
		sys.bcStack.PushI(c.mhv.uniqhit)
	default:
		sys.errLog.Printf("%v\n", be[*i-1])
		c.panic()
	}
}
func (be BytecodeExp) run_ex2(c *Char, i *int, oc *Char) {
	(*i)++
	switch be[*i-1] {
	case OC_ex2_index:
		sys.bcStack.PushI(c.index)
	case OC_ex2_runorder:
		sys.bcStack.PushI(c.runorder)
	case OC_ex2_palfxvar_time:
		sys.bcStack.PushI(c.palfxvar(0))
	case OC_ex2_palfxvar_addr:
		sys.bcStack.PushI(c.palfxvar(1))
	case OC_ex2_palfxvar_addg:
		sys.bcStack.PushI(c.palfxvar(2))
	case OC_ex2_palfxvar_addb:
		sys.bcStack.PushI(c.palfxvar(3))
	case OC_ex2_palfxvar_mulr:
		sys.bcStack.PushI(c.palfxvar(4))
	case OC_ex2_palfxvar_mulg:
		sys.bcStack.PushI(c.palfxvar(5))
	case OC_ex2_palfxvar_mulb:
		sys.bcStack.PushI(c.palfxvar(6))
	case OC_ex2_palfxvar_color:
		sys.bcStack.PushF(c.palfxvar2(1))
	case OC_ex2_palfxvar_hue:
		sys.bcStack.PushF(c.palfxvar2(2))
	case OC_ex2_palfxvar_invertall:
		sys.bcStack.PushI(c.palfxvar(-1))
	case OC_ex2_palfxvar_invertblend:
		sys.bcStack.PushI(c.palfxvar(-2))
	case OC_ex2_palfxvar_bg_time:
		sys.bcStack.PushI(sys.palfxvar(0, 1))
	case OC_ex2_palfxvar_bg_addr:
		sys.bcStack.PushI(sys.palfxvar(1, 1))
	case OC_ex2_palfxvar_bg_addg:
		sys.bcStack.PushI(sys.palfxvar(2, 1))
	case OC_ex2_palfxvar_bg_addb:
		sys.bcStack.PushI(sys.palfxvar(3, 1))
	case OC_ex2_palfxvar_bg_mulr:
		sys.bcStack.PushI(sys.palfxvar(4, 1))
	case OC_ex2_palfxvar_bg_mulg:
		sys.bcStack.PushI(sys.palfxvar(5, 1))
	case OC_ex2_palfxvar_bg_mulb:
		sys.bcStack.PushI(sys.palfxvar(6, 1))
	case OC_ex2_palfxvar_bg_color:
		sys.bcStack.PushF(sys.palfxvar2(1, 1))
	case OC_ex2_palfxvar_bg_hue:
		sys.bcStack.PushF(sys.palfxvar2(2, 1))
	case OC_ex2_palfxvar_bg_invertall:
		sys.bcStack.PushI(sys.palfxvar(-1, 1))
	case OC_ex2_palfxvar_all_time:
		sys.bcStack.PushI(sys.palfxvar(0, 2))
	case OC_ex2_palfxvar_all_addr:
		sys.bcStack.PushI(sys.palfxvar(1, 2))
	case OC_ex2_palfxvar_all_addg:
		sys.bcStack.PushI(sys.palfxvar(2, 2))
	case OC_ex2_palfxvar_all_addb:
		sys.bcStack.PushI(sys.palfxvar(3, 2))
	case OC_ex2_palfxvar_all_mulr:
		sys.bcStack.PushI(sys.palfxvar(4, 2))
	case OC_ex2_palfxvar_all_mulg:
		sys.bcStack.PushI(sys.palfxvar(5, 2))
	case OC_ex2_palfxvar_all_mulb:
		sys.bcStack.PushI(sys.palfxvar(6, 2))
	case OC_ex2_palfxvar_all_color:
		sys.bcStack.PushF(sys.palfxvar2(1, 2))
	case OC_ex2_palfxvar_all_hue:
		sys.bcStack.PushF(sys.palfxvar2(2, 2))
	case OC_ex2_palfxvar_all_invertall:
		sys.bcStack.PushI(sys.palfxvar(-1, 2))
	case OC_ex2_palfxvar_all_invertblend:
		sys.bcStack.PushI(sys.palfxvar(-2, 2))
	case OC_ex2_introstate:
		sys.bcStack.PushI(sys.introState())
	case OC_ex2_bgmvar_loopstart:
		if sys.bgm.volctrl != nil {
			if sl, ok := sys.bgm.volctrl.Streamer.(*StreamLooper); ok {
				sys.bcStack.PushI(int32(sl.loopstart))
			}
		}
	case OC_ex2_bgmvar_loopend:
		if sys.bgm.volctrl != nil {
			if sl, ok := sys.bgm.volctrl.Streamer.(*StreamLooper); ok {
				sys.bcStack.PushI(int32(sl.loopend))
			}
		}
	case OC_ex2_bgmvar_startposition:
		sys.bcStack.PushI(int32(sys.bgm.startPos))
	case OC_ex2_bgmvar_volume:
		sys.bcStack.PushI(int32(sys.bgm.bgmVolume))
	case OC_ex2_bgmvar_filename:
		sys.bcStack.PushB(sys.bgm.filename ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_ex2_gameoption_sound_bgmvolume:
		sys.bcStack.PushI(int32(sys.bgmVolume))
	case OC_ex2_gameoption_sound_mastervolume:
		sys.bcStack.PushI(int32(sys.masterVolume))
	case OC_ex2_gameoption_sound_maxvolume:
		sys.bcStack.PushI(int32(sys.maxBgmVolume))
	case OC_ex2_gameoption_sound_panningrange:
		sys.bcStack.PushI(int32(sys.panningRange))
	case OC_ex2_gameoption_sound_wavchannels:
		sys.bcStack.PushI(int32(sys.wavChannels))
	case OC_ex2_gameoption_sound_wavvolume:
		sys.bcStack.PushI(int32(sys.wavVolume))
	case OC_ex2_groundlevel:
		sys.bcStack.PushF(c.groundLevel)
	default:
		sys.errLog.Printf("%v\n", be[*i-1])
		c.panic()
	}
}
func (be BytecodeExp) evalF(c *Char) float32 {
	return be.run(c).ToF()
}
func (be BytecodeExp) evalI(c *Char) int32 {
	return be.run(c).ToI()
}
func (be BytecodeExp) evalI64(c *Char) int64 {
	return be.run(c).ToI64()
}
func (be BytecodeExp) evalB(c *Char) bool {
	return be.run(c).ToB()
}

type StateController interface {
	Run(c *Char, ps []int32) (changeState bool)
}
type NullStateController struct{}

func (NullStateController) Run(_ *Char, _ []int32) bool { return false }

var nullStateController NullStateController

type bytecodeFunction struct {
	numVars int32
	numRets int32
	numArgs int32
	ctrls   []StateController
}

func (bf bytecodeFunction) run(c *Char, ret []uint8) (changeState bool) {
	oldv, oldvslen := sys.bcVar, len(sys.bcVarStack)
	sys.bcVar = sys.bcVarStack.Alloc(int(bf.numVars))
	if len(sys.bcStack) != int(bf.numArgs) {
		c.panic()
	}
	copy(sys.bcVar, sys.bcStack)
	sys.bcStack.Clear()
	for _, sc := range bf.ctrls {
		switch sc.(type) {
		case StateBlock:
		default:
			if c.hitPause() {
				continue
			}
		}
		if sc.Run(c, nil) {
			changeState = true
			break
		}
	}
	if !changeState {
		if len(ret) > 0 {
			if len(ret) != int(bf.numRets) {
				c.panic()
			}
			for i, r := range ret {
				oldv[r] = sys.bcVar[int(bf.numArgs)+i]
			}
		}
	}
	sys.bcVar, sys.bcVarStack = oldv, sys.bcVarStack[:oldvslen]
	return
}

type callFunction struct {
	bytecodeFunction
	arg BytecodeExp
	ret []uint8
}

func (cf callFunction) Run(c *Char, _ []int32) (changeState bool) {
	if len(cf.arg) > 0 {
		sys.bcStack.Push(cf.arg.run(c))
	}
	return cf.run(c, cf.ret)
}

type StateBlock struct {
	// Basic block fields
	persistent          int32
	persistentIndex     int32
	ignorehitpause      int32
	ctrlsIgnorehitpause bool
	trigger             BytecodeExp
	elseBlock           *StateBlock
	ctrls               []StateController
	// Loop fields
	loopBlock        bool
	nestedInLoop     bool
	forLoop          bool
	forAssign        bool
	forCtrlVar       varAssign
	forExpression    [3]BytecodeExp
	forBegin, forEnd int32
	forIncrement     int32
}

func newStateBlock() *StateBlock {
	return &StateBlock{persistent: 1, persistentIndex: -1, ignorehitpause: -2}
}
func (b StateBlock) Run(c *Char, ps []int32) (changeState bool) {
	if c.hitPause() {
		if b.ignorehitpause < -1 {
			return false
		}
		if b.ignorehitpause >= 0 {
			ww := &c.ss.wakegawakaranai[sys.workingState.playerNo][b.ignorehitpause]
			*ww = !*ww
			if !*ww {
				return false
			}
		}
	}
	if b.persistentIndex >= 0 {
		ps[b.persistentIndex]--
		if ps[b.persistentIndex] > 0 {
			return false
		}
	}
	// https://github.com/ikemen-engine/Ikemen-GO/issues/963
	//sys.workingChar = c
	sys.workingChar = sys.chars[c.ss.sb.playerNo][0]
	if b.loopBlock {
		if b.forLoop {
			if b.forAssign {
				// Initial assign to control variable
				b.forCtrlVar.Run(c, ps)
				b.forBegin = sys.bcVar[b.forCtrlVar.vari].ToI()
			} else {
				b.forBegin = b.forExpression[0].evalI(c)
			}
			b.forEnd, b.forIncrement = b.forExpression[1].evalI(c), b.forExpression[2].evalI(c)
		}
		// Start loop
		interrupt := false
		for {
			// Decide if while loop should be stopped
			if !b.forLoop {
				// While loop needs to eval conditional indefinitely until it returns false
				if len(b.trigger) > 0 && !b.trigger.evalB(c) {
					interrupt = true
				}
			}
			// Run state controllers
			if !interrupt {
				for _, sc := range b.ctrls {
					switch sc.(type) {
					case StateBlock:
					default:
						if !b.ctrlsIgnorehitpause && c.hitPause() {
							continue
						}
					}
					if sc.Run(c, ps) {
						if sys.loopBreak {
							sys.loopBreak = false
							interrupt = true
							break
						}
						if sys.loopContinue {
							sys.loopContinue = false
							break
						}
						return true
					}
				}
			}
			// Decide if for loop should be stopped
			if b.forLoop {
				// Update loop count
				if b.forAssign {
					b.forBegin = sys.bcVar[b.forCtrlVar.vari].ToI() + b.forIncrement
				} else {
					b.forBegin += b.forIncrement
				}
				if b.forIncrement > 0 {
					if b.forBegin > b.forEnd {
						interrupt = true
					}
				} else if b.forBegin < b.forEnd {
					interrupt = true
				}
				// Update control variable if loop should keep going
				if b.forAssign && !interrupt {
					sys.bcVar[b.forCtrlVar.vari].SetI(b.forBegin)
				}
			}
			if interrupt {
				break
			}
		}
	} else {
		if len(b.trigger) > 0 && !b.trigger.evalB(c) {
			if b.elseBlock != nil {
				return b.elseBlock.Run(c, ps)
			}
			return false
		}
		for _, sc := range b.ctrls {
			switch sc.(type) {
			case StateBlock:
			default:
				if !b.ctrlsIgnorehitpause && c.hitPause() {
					continue
				}
			}
			if sc.Run(c, ps) {
				return true
			}
		}
	}
	if b.persistentIndex >= 0 {
		ps[b.persistentIndex] = b.persistent
	}
	return false
}

type StateExpr BytecodeExp

func (se StateExpr) Run(c *Char, _ []int32) (changeState bool) {
	BytecodeExp(se).run(c)
	return false
}

type varAssign struct {
	vari uint8
	be   BytecodeExp
}

func (va varAssign) Run(c *Char, _ []int32) (changeState bool) {
	sys.bcVar[va.vari] = va.be.run(c)
	return false
}

type LoopBreak struct{}

func (lb LoopBreak) Run(c *Char, _ []int32) (stop bool) {
	sys.loopBreak = true
	return true
}

type LoopContinue struct{}

func (lc LoopContinue) Run(c *Char, _ []int32) (stop bool) {
	sys.loopContinue = true
	return true
}

type StateControllerBase []byte

func newStateControllerBase() *StateControllerBase {
	return (*StateControllerBase)(&[]byte{})
}
func (StateControllerBase) beToExp(be ...BytecodeExp) []BytecodeExp {
	return be
}
func (StateControllerBase) fToExp(f ...float32) (exp []BytecodeExp) {
	for _, v := range f {
		var be BytecodeExp
		be.appendValue(BytecodeFloat(v))
		exp = append(exp, be)
	}
	return
}
func (StateControllerBase) iToExp(i ...int32) (exp []BytecodeExp) {
	for _, v := range i {
		var be BytecodeExp
		be.appendValue(BytecodeInt(v))
		exp = append(exp, be)
	}
	return
}
func (StateControllerBase) i64ToExp(i ...int64) (exp []BytecodeExp) {
	for _, v := range i {
		var be BytecodeExp
		be.appendValue(BytecodeInt64(v))
		exp = append(exp, be)
	}
	return
}
func (StateControllerBase) bToExp(i bool) (exp []BytecodeExp) {
	var be BytecodeExp
	be.appendValue(BytecodeBool(i))
	exp = append(exp, be)
	return
}
func (scb *StateControllerBase) add(id byte, exp []BytecodeExp) {
	*scb = append(*scb, id, byte(len(exp)))
	for _, e := range exp {
		l := int32(len(e))
		*scb = append(*scb, (*(*[4]byte)(unsafe.Pointer(&l)))[:]...)
		*scb = append(*scb, *(*[]byte)(unsafe.Pointer(&e))...)
	}
}
func (scb StateControllerBase) run(c *Char,
	f func(byte, []BytecodeExp) bool) {
	for i := 0; i < len(scb); {
		id := scb[i]
		i++
		n := scb[i]
		i++
		if cap(sys.workBe) < int(n) {
			sys.workBe = make([]BytecodeExp, n)
		} else {
			sys.workBe = sys.workBe[:n]
		}
		for m := 0; m < int(n); m++ {
			l := *(*int32)(unsafe.Pointer(&scb[i]))
			i += 4
			sys.workBe[m] = (*(*BytecodeExp)(unsafe.Pointer(&scb)))[i : i+int(l)]
			i += int(l)
		}
		if !f(id, sys.workBe) {
			break
		}
	}
}

type stateDef StateControllerBase

const (
	stateDef_hitcountpersist byte = iota
	stateDef_movehitpersist
	stateDef_hitdefpersist
	stateDef_sprpriority
	stateDef_facep2
	stateDef_juggle
	stateDef_velset
	stateDef_anim
	stateDef_ctrl
	stateDef_poweradd
)

func (sc stateDef) Run(c *Char) {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case stateDef_hitcountpersist:
			if !exp[0].evalB(c) {
				c.clearHitCount()
			}
		case stateDef_movehitpersist:
			if !exp[0].evalB(c) {
				c.clearMoveHit()
			}
		case stateDef_hitdefpersist:
			if !exp[0].evalB(c) {
				c.clearHitDef()
			}
		case stateDef_sprpriority:
			c.setSprPriority(exp[0].evalI(c))
		case stateDef_facep2:
			if exp[0].evalB(c) && c.rdDistX(c.p2(), c).ToF() < 0 {
				c.setFacing(-c.facing)
			}
		case stateDef_juggle:
			c.setJuggle(exp[0].evalI(c))
		case stateDef_velset:
			c.setXV(exp[0].evalF(c))
			if len(exp) > 1 {
				c.setYV(exp[1].evalF(c))
				if len(exp) > 2 {
					exp[2].run(c)
				}
			}
		case stateDef_anim:
			c.changeAnimEx(exp[1].evalI(c), c.playerNo, string(*(*[]byte)(unsafe.Pointer(&exp[0]))), false)
		case stateDef_ctrl:
			//in mugen fatal blow ignores statedef ctrl
			if !c.ghv.fatal {
				c.setCtrl(exp[0].evalB(c))
			} else {
				c.ghv.fatal = false
			}
		case stateDef_poweradd:
			c.powerAdd(exp[0].evalI(c))
		}
		return true
	})
}

type hitBy StateControllerBase

const (
	hitBy_value byte = iota
	hitBy_value2
	hitBy_time
	hitBy_redirectid
)

func (sc hitBy) Run(c *Char, _ []int32) bool {
	time := int32(1)
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case hitBy_time:
			time = exp[0].evalI(c)
		case hitBy_value:
			crun.hitby[0].time = time
			crun.hitby[0].flag = exp[0].evalI(c)
		case hitBy_value2:
			crun.hitby[1].time = time
			crun.hitby[1].flag = exp[0].evalI(c)
		case hitBy_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type notHitBy hitBy

func (sc notHitBy) Run(c *Char, _ []int32) bool {
	time := int32(1)
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case hitBy_time:
			time = exp[0].evalI(c)
		case hitBy_value:
			crun.hitby[0].time = time
			crun.hitby[0].flag = ^exp[0].evalI(c)
		case hitBy_value2:
			crun.hitby[1].time = time
			crun.hitby[1].flag = ^exp[0].evalI(c)

		case hitBy_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type assertSpecial StateControllerBase

const (
	assertSpecial_flag byte = iota
	assertSpecial_flag_g
	assertSpecial_noko
	assertSpecial_redirectid
)

func (sc assertSpecial) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case assertSpecial_flag:
			crun.setASF(AssertSpecialFlag(exp[0].evalI64(c)))
		case assertSpecial_flag_g:
			sys.setGSF(GlobalSpecialFlag(exp[0].evalI(c)))
		case assertSpecial_noko:
			if c.stWgi().ikemenver[0] != 0 || c.stWgi().ikemenver[1] != 0 {
				crun.setASF(AssertSpecialFlag(ASF_noko))
			} else {
				sys.setGSF(GlobalSpecialFlag(GSF_noko))
			}
		case assertSpecial_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type playSnd StateControllerBase

const (
	playSnd_value = iota
	playSnd_channel
	playSnd_lowpriority
	playSnd_pan
	playSnd_abspan
	playSnd_volume
	playSnd_volumescale
	playSnd_freqmul
	playSnd_loop
	playSnd_redirectid
	playSnd_priority
	playSnd_loopstart
	playSnd_loopend
	playSnd_startposition
	playSnd_loopcount
	playSnd_stopongethit
	playSnd_stoponchangestate
)

func (sc playSnd) Run(c *Char, _ []int32) bool {
	if sys.noSoundFlg {
		return false
	}
	crun := c
	f, lw, lp, stopgh, stopcs := "", false, false, false, false
	var g, n, ch, vo, pri, lc int32 = -1, 0, -1, 100, 0, 0
	var loopstart, loopend, startposition = 0, 0, 0
	var p, fr float32 = 0, 1
	x := &c.pos[0]
	ls := c.localscl
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case playSnd_value:
			f = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			g = exp[1].evalI(c)
			if len(exp) > 2 {
				n = exp[2].evalI(c)
			}
		case playSnd_channel:
			ch = exp[0].evalI(c)
			if ch == 0 {
				stopgh = true
			}
		case playSnd_lowpriority:
			lw = exp[0].evalB(c)
		case playSnd_pan:
			p = exp[0].evalF(c)
		case playSnd_abspan:
			x = nil
			ls = 1
			p = exp[0].evalF(c)
		case playSnd_volume:
			vo = vo + int32(float64(exp[0].evalI(c))*(25.0/64.0))
		case playSnd_volumescale:
			vo = exp[0].evalI(c)
		case playSnd_freqmul:
			fr = ClampF(exp[0].evalF(c), 0.01, 5)
		case playSnd_loop:
			lp = exp[0].evalB(c)
		case playSnd_priority:
			pri = exp[0].evalI(c)
		case playSnd_loopstart:
			loopstart = int(exp[0].evalI64(c))
		case playSnd_loopend:
			loopend = int(exp[0].evalI64(c))
		case playSnd_startposition:
			startposition = int(exp[0].evalI64(c))
		case playSnd_loopcount:
			lc = exp[0].evalI(c)
		case playSnd_stopongethit:
			stopgh = exp[0].evalB(c)
		case playSnd_stoponchangestate:
			stopcs = exp[0].evalB(c)
		case playSnd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				x = &crun.pos[0]
				ls = crun.localscl
			} else {
				return false
			}
		}
		return true
	})
	// Read the loop parameter if loopcount not specified
	if lc == 0 {
		if lp {
			crun.playSound(f, lw, -1, g, n, ch, vo, p, fr, ls, x, true, pri, loopstart, loopend, startposition, stopgh, stopcs)
		} else {
			crun.playSound(f, lw, 0, g, n, ch, vo, p, fr, ls, x, true, pri, loopstart, loopend, startposition, stopgh, stopcs)
		}
		// Use the loopcount directly if it's been specified
	} else {
		crun.playSound(f, lw, lc, g, n, ch, vo, p, fr, ls, x, true, pri, loopstart, loopend, startposition, stopgh, stopcs)
	}
	return false
}

type changeState StateControllerBase

const (
	changeState_value byte = iota
	changeState_ctrl
	changeState_anim
	changeState_continue
	changeState_readplayerid
	changeState_redirectid
)

func (sc changeState) Run(c *Char, _ []int32) bool {
	crun := c
	var v, a, ctrl int32 = -1, -1, -1
	ffx := ""
	stop := true
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case changeState_value:
			v = exp[0].evalI(c)
		case changeState_ctrl:
			ctrl = exp[0].evalI(c)
		case changeState_anim:
			a = exp[1].evalI(c)
			ffx = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		case changeState_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				stop = false
				crun = rid
			} else {
				return false
			}
		case changeState_continue:
			stop = !exp[0].evalB(c)
		}
		return true
	})
	crun.changeState(v, a, ctrl, ffx)
	return stop
}

type selfState changeState

func (sc selfState) Run(c *Char, _ []int32) bool {
	crun := c
	var v, a, r, ctrl int32 = -1, -1, -1, -1
	ffx := ""
	stop := true
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case changeState_value:
			v = exp[0].evalI(c)
		case changeState_ctrl:
			ctrl = exp[0].evalI(c)
		case changeState_anim:
			a = exp[1].evalI(c)
			ffx = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		case changeState_readplayerid:
			if rpid := sys.playerID(exp[0].evalI(c)); rpid != nil {
				r = int32(rpid.playerNo)
			} else {
				return false
			}
		case changeState_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				stop = false
				crun = rid
			} else {
				return false
			}
		case changeState_continue:
			stop = !exp[0].evalB(c)
		}
		return true
	})
	crun.selfState(v, a, r, ctrl, ffx)
	return stop
}

type tagIn StateControllerBase

const (
	tagIn_stateno = iota
	tagIn_partnerstateno
	tagIn_self
	tagIn_partner
	tagIn_ctrl
	tagIn_partnerctrl
	tagIn_leader
	tagIn_redirectid
)

func (sc tagIn) Run(c *Char, _ []int32) bool {
	crun := c
	var tagSCF int32 = -1
	var partnerNo int32 = -1
	var partnerStateNo int32 = -1
	var partnerCtrlSetting int32 = -1
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case tagIn_stateno:
			sn := exp[0].evalI(c)
			if sn >= 0 {
				crun.changeState(sn, -1, -1, "")
				if tagSCF == -1 {
					tagSCF = 1
				}
			} else {
				return false
			}
		case tagIn_partnerstateno:
			if psn := exp[0].evalI(c); psn >= 0 {
				partnerStateNo = psn
			} else {
				return false
			}
		case tagIn_self:
			tagSCF = Btoi(exp[0].evalB(c))
		case tagIn_partner:
			pti := exp[0].evalI(c)
			if pti >= 0 {
				partnerNo = pti
			} else {
				return false
			}
		case tagIn_ctrl:
			ctrls := exp[0].evalB(c)
			crun.setCtrl(ctrls)
			if tagSCF == -1 {
				tagSCF = 1
			}
		case tagIn_partnerctrl:
			partnerCtrlSetting = Btoi(exp[0].evalB(c))
		case tagIn_leader:
			if crun.teamside != -1 {
				ld := int(exp[0].evalI(c)) - 1
				if ld&1 == crun.playerNo&1 && ld >= crun.teamside && ld <= int(sys.numSimul[crun.teamside])*2-^crun.teamside&1-1 {
					sys.teamLeader[crun.playerNo&1] = ld
				}
			}
		case tagIn_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	// Data adjustments
	if tagSCF == -1 && partnerNo == -1 {
		tagSCF = 1
	}
	if tagSCF == 1 {
		crun.unsetSCF(SCF_standby)
	}
	// Partner
	if partnerNo != -1 && crun.partnerV2(partnerNo) != nil {
		partner := crun.partnerV2(partnerNo)
		partner.unsetSCF(SCF_standby)
		if partnerStateNo >= 0 {
			partner.changeState(partnerStateNo, -1, -1, "")
		}
		if partnerCtrlSetting != -1 {
			if partnerCtrlSetting == 1 {
				partner.setCtrl(true)
			} else {
				partner.setCtrl(false)
			}
		}
	}
	return false
}

type tagOut StateControllerBase

const (
	tagOut_self = iota
	tagOut_partner
	tagOut_stateno
	tagOut_partnerstateno
	tagOut_redirectid
)

func (sc tagOut) Run(c *Char, _ []int32) bool {
	crun := c
	var tagSCF int32 = -1
	var partnerNo int32 = -1
	var partnerStateNo int32 = -1
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case tagOut_self:
			tagSCF = Btoi(exp[0].evalB(c))
		case tagOut_stateno:
			sn := exp[0].evalI(c)
			if sn >= 0 {
				crun.changeState(sn, -1, -1, "")
				if tagSCF == -1 {
					tagSCF = 1
				}
			} else {
				return false
			}
		case tagOut_partner:
			pti := exp[0].evalI(c)
			if pti >= 0 {
				partnerNo = pti
			} else {
				return false
			}
		case tagOut_partnerstateno:
			if psn := exp[0].evalI(c); psn >= 0 {
				partnerStateNo = psn
			} else {
				return false
			}
		case tagOut_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	if tagSCF == -1 && partnerNo == -1 && partnerStateNo == -1 {
		tagSCF = 1
	}
	if tagSCF == 1 {
		crun.setSCF(SCF_standby)
		sys.charList.p2enemyDelete(crun)
	}
	if partnerNo != -1 && crun.partnerV2(partnerNo) != nil {
		partner := crun.partnerV2(partnerNo)
		partner.setSCF(SCF_standby)
		if partnerStateNo >= 0 {
			partner.changeState(partnerStateNo, -1, -1, "")
		}
		sys.charList.p2enemyDelete(partner)
	}
	return false
}

type destroySelf StateControllerBase

const (
	destroySelf_recursive = iota
	destroySelf_removeexplods
	destroySelf_redirectid
)

func (sc destroySelf) Run(c *Char, _ []int32) bool {
	crun := c
	rec, rem := false, false
	self := true
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case destroySelf_recursive:
			rec = exp[0].evalB(c)
		case destroySelf_removeexplods:
			rem = exp[0].evalB(c)
		case destroySelf_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				self = rid.id == c.id
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return crun.destroySelf(rec, rem) && self
}

type changeAnim StateControllerBase

const (
	changeAnim_elem byte = iota
	changeAnim_value
	changeAnim_readplayerid
	changeAnim_redirectid
)

func (sc changeAnim) Run(c *Char, _ []int32) bool {
	crun := c
	var elem int32
	var r int = -1
	setelem := false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case changeAnim_elem:
			elem = exp[0].evalI(c)
			setelem = true
		case changeAnim_value:
			pn := crun.playerNo
			if r != -1 {
				pn = r
			}
			crun.changeAnim(exp[1].evalI(c), pn, string(*(*[]byte)(unsafe.Pointer(&exp[0]))))
			if setelem {
				crun.setAnimElem(elem)
			}
		case changeAnim_readplayerid:
			if rpid := sys.playerID(exp[0].evalI(c)); rpid != nil {
				r = rpid.playerNo
			} else {
				return false
			}
		case changeAnim_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type changeAnim2 changeAnim

func (sc changeAnim2) Run(c *Char, _ []int32) bool {
	crun := c
	var elem int32
	setelem := false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case changeAnim_elem:
			elem = exp[0].evalI(c)
			setelem = true
		case changeAnim_value:
			crun.changeAnim2(exp[1].evalI(c), string(*(*[]byte)(unsafe.Pointer(&exp[0]))))
			if setelem {
				crun.setAnimElem(elem)
			}
		case changeAnim_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type helper StateControllerBase

const (
	helper_helpertype byte = iota
	helper_name
	helper_postype
	helper_ownpal
	helper_size_xscale
	helper_size_yscale
	helper_size_ground_back
	helper_size_ground_front
	helper_size_air_back
	helper_size_air_front
	helper_size_height_stand
	helper_size_height_crouch
	helper_size_height_air
	helper_size_height_down
	helper_size_proj_doscale
	helper_size_head_pos
	helper_size_mid_pos
	helper_size_shadowoffset
	helper_stateno
	helper_keyctrl
	helper_id
	helper_pos
	helper_facing
	helper_pausemovetime
	helper_supermovetime
	helper_redirectid
	helper_remappal
	helper_extendsmap
	helper_inheritjuggle
	helper_inheritchannels
	helper_immortal
	helper_kovelocity
	helper_preserve
	helper_standby
)

func (sc helper) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	var h *Char
	pt := PT_P1
	var f, st int32 = 1, 0
	var extmap bool
	var x, y float32 = 0, 0
	rp := [...]int32{-1, 0}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if h == nil {
			if id == helper_redirectid {
				if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
					crun = rid
					lclscround = c.localscl / crun.localscl
					h = crun.newHelper()
				} else {
					return false
				}
			} else {
				h = c.newHelper()
			}
		}
		if h == nil {
			return false
		}
		switch id {
		case helper_helpertype:
			h.player = exp[0].evalB(c)
		case helper_name:
			h.name = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		case helper_postype:
			pt = PosType(exp[0].evalI(c))
		case helper_ownpal:
			h.ownpal = exp[0].evalB(c)
		case helper_size_xscale:
			h.size.xscale = exp[0].evalF(c)
		case helper_size_yscale:
			h.size.yscale = exp[0].evalF(c)
		case helper_size_ground_back:
			h.size.ground.back = exp[0].evalF(c)
		case helper_size_ground_front:
			h.size.ground.front = exp[0].evalF(c)
		case helper_size_air_back:
			h.size.air.back = exp[0].evalF(c)
		case helper_size_air_front:
			h.size.air.front = exp[0].evalF(c)
		case helper_size_height_stand:
			h.size.height.stand = exp[0].evalF(c)
		case helper_size_height_crouch:
			h.size.height.crouch = exp[0].evalF(c)
		case helper_size_height_air:
			h.size.height.air[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				h.size.height.air[1] = exp[1].evalF(c)
			}
		case helper_size_height_down:
			h.size.height.down = exp[0].evalF(c)
		case helper_size_proj_doscale:
			h.size.proj.doscale = exp[0].evalI(c)
		case helper_size_head_pos:
			h.size.head.pos[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				h.size.head.pos[1] = exp[1].evalF(c)
			}
		case helper_size_mid_pos:
			h.size.mid.pos[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				h.size.mid.pos[1] = exp[1].evalF(c)
			}
		case helper_size_shadowoffset:
			h.size.shadowoffset = exp[0].evalF(c)
		case helper_stateno:
			st = exp[0].evalI(c)
		case helper_keyctrl:
			for _, e := range exp {
				m := e.run(c).ToI()
				if m > 0 && m <= int32(len(h.keyctrl)) {
					h.keyctrl[m-1] = true
				}
			}
		case helper_id:
			h.helperId = exp[0].evalI(c)
		case helper_pos:
			x = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				y = exp[1].evalF(c) * lclscround
			}
		case helper_facing:
			f = exp[0].evalI(c)
		case helper_pausemovetime:
			h.pauseMovetime = exp[0].evalI(c)
		case helper_supermovetime:
			h.superMovetime = exp[0].evalI(c)
		case helper_remappal:
			rp[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				rp[1] = exp[1].evalI(c)
			}
		case helper_extendsmap:
			extmap = exp[0].evalB(c)
		case helper_inheritjuggle:
			h.inheritJuggle = exp[0].evalI(c)
		case helper_inheritchannels:
			h.inheritChannels = exp[0].evalI(c)
		case helper_immortal:
			h.immortal = exp[0].evalB(c)
		case helper_kovelocity:
			h.kovelocity = exp[0].evalB(c)
		case helper_preserve:
			if exp[0].evalB(c) {
				h.preserve = sys.round
			}
		case helper_standby:
			if exp[0].evalB(c) {
				h.setSCF(SCF_standby)
			} else {
				h.unsetSCF(SCF_standby)
			}
		}
		return true
	})
	if h == nil {
		return false
	}
	if crun.minus == -2 || crun.minus == -4 {
		h.localscl = (320 / crun.localcoord)
		h.localcoord = crun.localcoord
	} else {
		h.localscl = crun.localscl
		h.localcoord = crun.localcoord
	}
	crun.helperInit(h, st, pt, x, y, f, rp, extmap)
	return false
}

type ctrlSet StateControllerBase

const (
	ctrlSet_value byte = iota
	ctrlSet_redirectid
)

func (sc ctrlSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case ctrlSet_value:
			crun.setCtrl(exp[0].evalB(c))
		case ctrlSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type posSet StateControllerBase

const (
	posSet_x byte = iota
	posSet_y
	posSet_z
	posSet_redirectid
)

func (sc posSet) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case posSet_x:
			x := sys.cam.Pos[0]/crun.localscl + exp[0].evalF(c)*lclscround
			crun.setX(x)
			if crun.bindToId > 0 && !math.IsNaN(float64(crun.bindPos[0])) && sys.playerID(crun.bindToId) != nil {
				crun.bindPosAdd[0] = x
			}
		case posSet_y:
			y := exp[0].evalF(c)*lclscround + crun.groundLevel + crun.platformPosY
			crun.setY(y)
			if crun.bindToId > 0 && !math.IsNaN(float64(crun.bindPos[1])) && sys.playerID(crun.bindToId) != nil {
				crun.bindPosAdd[1] = y
			}
		case posSet_z:
			if crun.size.z.enable {
				crun.setZ(exp[0].evalF(c) * lclscround)
			} else {
				exp[0].run(c)
			}
		case posSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type posAdd posSet

func (sc posAdd) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case posSet_x:
			x := exp[0].evalF(c) * lclscround
			crun.addX(x)
			if crun.bindToId > 0 && !math.IsNaN(float64(crun.bindPos[0])) && sys.playerID(crun.bindToId) != nil {
				crun.bindPosAdd[0] = x
			}
		case posSet_y:
			y := exp[0].evalF(c) * lclscround
			crun.addY(y)
			if crun.bindToId > 0 && !math.IsNaN(float64(crun.bindPos[1])) && sys.playerID(crun.bindToId) != nil {
				crun.bindPosAdd[1] = y
			}
		case posSet_z:
			if crun.size.z.enable {
				crun.addZ(exp[0].evalF(c) * lclscround)
			} else {
				exp[0].run(c)
			}
		case posSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type velSet posSet

func (sc velSet) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case posSet_x:
			crun.setXV(exp[0].evalF(c) * lclscround)
		case posSet_y:
			crun.setYV(exp[0].evalF(c) * lclscround)
		case posSet_z:
			if crun.size.z.enable {
				crun.setZV(exp[0].evalF(c) * lclscround)
			} else {
				exp[0].run(c)
			}
		case posSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type velAdd posSet

func (sc velAdd) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case posSet_x:
			crun.addXV(exp[0].evalF(c) * lclscround)
		case posSet_y:
			crun.addYV(exp[0].evalF(c) * lclscround)
		case posSet_z:
			if crun.size.z.enable {
				crun.addZV(exp[0].evalF(c) * lclscround)
			} else {
				exp[0].run(c)
			}
		case posSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type velMul posSet

func (sc velMul) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case posSet_x:
			crun.mulXV(exp[0].evalF(c))
		case posSet_y:
			crun.mulYV(exp[0].evalF(c))
		case posSet_z:
			if crun.size.z.enable {
				crun.mulZV(exp[0].evalF(c))
			} else {
				exp[0].run(c)
			}
		case posSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type palFX StateControllerBase

const (
	palFX_time byte = iota
	palFX_color
	palFX_add
	palFX_mul
	palFX_sinadd
	palFX_sinmul
	palFX_sincolor
	palFX_sinhue
	palFX_invertall
	palFX_invertblend
	palFX_hue
	palFX_last = iota - 1
	palFX_redirectid
)

func (sc palFX) runSub(c *Char, pfd *PalFXDef,
	id byte, exp []BytecodeExp) bool {
	switch id {
	case palFX_time:
		pfd.time = exp[0].evalI(c)
	case palFX_color:
		pfd.color = exp[0].evalF(c) / 256
	case palFX_hue:
		pfd.hue = exp[0].evalF(c) / 256
	case palFX_add:
		pfd.add[0] = exp[0].evalI(c)
		pfd.add[1] = exp[1].evalI(c)
		pfd.add[2] = exp[2].evalI(c)
	case palFX_mul:
		pfd.mul[0] = exp[0].evalI(c)
		pfd.mul[1] = exp[1].evalI(c)
		pfd.mul[2] = exp[2].evalI(c)
	case palFX_sinadd:
		var side int32 = 1
		if len(exp) > 3 {
			if exp[3].evalI(c) < 0 {
				pfd.cycletime[0] = -exp[3].evalI(c)
				side = -1
			} else {
				pfd.cycletime[0] = exp[3].evalI(c)
			}
		}
		pfd.sinadd[0] = exp[0].evalI(c) * side
		pfd.sinadd[1] = exp[1].evalI(c) * side
		pfd.sinadd[2] = exp[2].evalI(c) * side
	case palFX_sinmul:
		var side int32 = 1
		if len(exp) > 3 {
			if exp[3].evalI(c) < 0 {
				pfd.cycletime[1] = -exp[3].evalI(c)
				side = -1
			} else {
				pfd.cycletime[1] = exp[3].evalI(c)
			}
		}
		pfd.sinmul[0] = exp[0].evalI(c) * side
		pfd.sinmul[1] = exp[1].evalI(c) * side
		pfd.sinmul[2] = exp[2].evalI(c) * side
	case palFX_sincolor:
		var side int32 = 1
		if len(exp) > 1 {
			if exp[1].evalI(c) < 0 {
				pfd.cycletime[2] = -exp[1].evalI(c)
				side = -1
			} else {
				pfd.cycletime[2] = exp[1].evalI(c)
			}
		}
		pfd.sincolor = exp[0].evalI(c) * side
	case palFX_sinhue:
		var side int32 = 1
		if len(exp) > 1 {
			if exp[1].evalI(c) < 0 {
				pfd.cycletime[3] = -exp[1].evalI(c)
				side = -1
			} else {
				pfd.cycletime[3] = exp[1].evalI(c)
			}
		}
		pfd.sinhue = exp[0].evalI(c) * side
	case palFX_invertall:
		pfd.invertall = exp[0].evalB(c)
	case palFX_invertblend:
		pfd.invertblend = Clamp(exp[0].evalI(c), -1, 2)
	default:
		return false
	}
	return true
}
func (sc palFX) Run(c *Char, _ []int32) bool {
	crun := c
	doOnce := false
	pf := newPalFX()
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if id == palFX_redirectid {
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		if !doOnce {
			if !crun.ownpal {
				return false
			}
			pf = crun.palfx
			if pf == nil {
				pf = newPalFX()
			}
			pf.clear2(true)
			//Mugen 1.1 behavior if invertblend param is omitted (Only if char mugenversion = 1.1)
			if c.stWgi().mugenver[0] == 1 && c.stWgi().mugenver[1] == 1 && c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
				pf.invertblend = -2
			}
			doOnce = true
		}
		sc.runSub(c, &pf.PalFXDef, id, exp)
		return true
	})
	return false
}

type allPalFX palFX

func (sc allPalFX) Run(c *Char, _ []int32) bool {
	sys.allPalFX.clear()
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		palFX(sc).runSub(c, &sys.allPalFX.PalFXDef, id, exp)
		//Forcing 1.1 kind behavior
		sys.allPalFX.invertblend = Clamp(sys.allPalFX.invertblend, 0, 1)
		return true
	})
	return false
}

type bgPalFX palFX

func (sc bgPalFX) Run(c *Char, _ []int32) bool {
	sys.bgPalFX.clear()
	//Forcing 1.1 behavior
	sys.bgPalFX.invertblend = -2
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		palFX(sc).runSub(c, &sys.bgPalFX.PalFXDef, id, exp)
		sys.bgPalFX.invertblend = -3
		return true
	})
	return false
}

type explod StateControllerBase

const (
	explod_anim byte = iota + palFX_last + 1
	explod_ownpal
	explod_remappal
	explod_id
	explod_facing
	explod_vfacing
	explod_pos
	explod_random
	explod_postype
	explod_velocity
	explod_accel
	explod_scale
	explod_bindtime
	explod_removetime
	explod_supermove
	explod_supermovetime
	explod_pausemovetime
	explod_sprpriority
	explod_ontop
	explod_strictontop
	explod_under
	explod_shadow
	explod_removeongethit
	explod_removeonchangestate
	explod_trans
	explod_animelem
	explod_animfreeze
	explod_angle
	explod_yangle
	explod_xangle
	explod_projection
	explod_focallength
	explod_ignorehitpause
	explod_bindid
	explod_space
	explod_window
	explod_postypeExists
	explod_interpolate_time
	explod_interpolate_animelem
	explod_interpolate_pos
	explod_interpolate_scale
	explod_interpolate_angle
	explod_interpolate_alpha
	explod_interpolate_focallength
	explod_interpolate_pfx_mul
	explod_interpolate_pfx_add
	explod_interpolate_pfx_color
	explod_interpolate_pfx_hue
	explod_interpolation
	explod_redirectid
)

func (sc explod) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	var e *Explod
	var i int
	//e, i := crun.newExplod()
	rp := [...]int32{-1, 0}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if e == nil {
			if id == explod_redirectid {
				if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
					crun = rid
					lclscround = c.localscl / crun.localscl
					e, i = crun.newExplod()
					if e == nil {
						return false
					}
					e.id = 0
				} else {
					return false
				}
			} else {
				e, i = crun.newExplod()
				if e == nil {
					return false
				}
				e.id = 0
			}
			// Mugenversion 1.1 chars default postype to "None"
			if c.stWgi().mugenver[0] == 1 && c.stWgi().mugenver[1] == 1 {
				e.postype = PT_None
			}
		}
		switch id {
		case explod_anim:
			ffx := string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			if ffx != "" && ffx != "s" {
				e.ownpal = true
			}
			e.anim = crun.getAnim(exp[1].evalI(c), ffx, true)
		case explod_ownpal:
			e.ownpal = exp[0].evalB(c)
		case explod_remappal:
			rp[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				rp[1] = exp[1].evalI(c)
			}
		case explod_id:
			e.id = Max(0, exp[0].evalI(c))
		case explod_facing:
			if exp[0].evalI(c) < 0 {
				e.relativef = -1
			} else {
				e.relativef = 1
			}
		case explod_vfacing:
			if exp[0].evalI(c) < 0 {
				e.vfacing = -1
			} else {
				e.vfacing = 1
			}
		case explod_pos:
			e.relativePos[0] = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				e.relativePos[1] = exp[1].evalF(c) * lclscround
			}
		case explod_random:
			rndx := (exp[0].evalF(c) / 2) * lclscround
			e.relativePos[0] += RandF(-rndx, rndx)
			if len(exp) > 1 {
				rndy := (exp[1].evalF(c) / 2) * lclscround
				e.relativePos[1] += RandF(-rndy, rndy)
			}
		case explod_space:
			e.space = Space(exp[0].evalI(c))
		case explod_postype:
			e.postype = PosType(exp[0].evalI(c))
		case explod_velocity:
			e.velocity[0] = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				e.velocity[1] = exp[1].evalF(c) * lclscround
			}
		case explod_accel:
			e.accel[0] = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				e.accel[1] = exp[1].evalF(c) * lclscround
			}
		case explod_scale:
			e.scale[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				e.scale[1] = exp[1].evalF(c)
			}
		case explod_bindtime:
			e.bindtime = exp[0].evalI(c)
		case explod_removetime:
			e.removetime = exp[0].evalI(c)
		case explod_supermove:
			if exp[0].evalB(c) {
				e.supermovetime = -1
			} else {
				e.supermovetime = 0
			}
		case explod_supermovetime:
			e.supermovetime = exp[0].evalI(c)
			if e.supermovetime >= 0 {
				e.supermovetime = Max(e.supermovetime, e.supermovetime+1)
			}
		case explod_pausemovetime:
			e.pausemovetime = exp[0].evalI(c)
			if e.pausemovetime >= 0 {
				e.pausemovetime = Max(e.pausemovetime, e.pausemovetime+1)
			}
		case explod_sprpriority:
			e.sprpriority = exp[0].evalI(c)
		case explod_ontop:
			e.ontop = exp[0].evalB(c)
		case explod_strictontop:
			if e.ontop {
				e.sprpriority = 0
			}
		case explod_under:
			if !e.ontop {
				e.under = exp[0].evalB(c)
			}
		case explod_shadow:
			e.shadow[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				e.shadow[1] = exp[1].evalI(c)
				if len(exp) > 2 {
					e.shadow[2] = exp[2].evalI(c)
				}
			}
		case explod_removeongethit:
			e.removeongethit = exp[0].evalB(c)
		case explod_removeonchangestate:
			e.removeonchangestate = exp[0].evalB(c)
		case explod_trans:
			e.alpha[0] = exp[0].evalI(c)
			e.alpha[1] = exp[1].evalI(c)
			sa, da := e.alpha[0], e.alpha[1]

			if len(exp) >= 3 {
				e.alpha[0] = Clamp(e.alpha[0], 0, 255)
				e.alpha[1] = Clamp(e.alpha[1], 0, 255)
				//if len(exp) >= 4 {
				//	e.alpha[1] = ^e.alpha[1]
				//} else if e.alpha[0] == 1 && e.alpha[1] == 255 {

				//Add
				e.blendmode = 1
				//Sub
				if sa == 1 && da == 255 {
					e.blendmode = 2
				} else if sa == -1 && da == 0 {
					e.blendmode = 0
				}
				if e.alpha[0] == 1 && e.alpha[1] == 255 {
					e.alpha[0] = 0
				}
			}
		case explod_animelem:
			animelem := exp[0].evalI(c)
			e.animelem = animelem
			e.anim.Action()
			e.setAnimElem()
		case explod_animfreeze:
			e.animfreeze = exp[0].evalB(c)
		case explod_angle:
			e.anglerot[0] = exp[0].evalF(c)
		case explod_yangle:
			e.anglerot[2] = exp[0].evalF(c)
		case explod_xangle:
			e.anglerot[1] = exp[0].evalF(c)
		case explod_focallength:
			e.fLength = exp[0].evalF(c)
		case explod_ignorehitpause:
			e.ignorehitpause = exp[0].evalB(c)
		case explod_bindid:
			bId := exp[0].evalI(c)
			if bId == -1 {
				bId = crun.id
			}
			e.setBind(bId)
		case explod_projection:
			e.projection = Projection(exp[0].evalI(c))
		case explod_window:
			e.window = [4]float32{exp[0].evalF(c) * lclscround, exp[1].evalF(c) * lclscround, exp[2].evalF(c) * lclscround, exp[3].evalF(c) * lclscround}
		default:
			if c.stWgi().mugenver[0] == 1 && c.stWgi().mugenver[1] == 1 && c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
				e.palfxdef.invertblend = -2
			}
			palFX(sc).runSub(c, &e.palfxdef, id, exp)

			explod(sc).setInterpolation(c, e, id, exp, &e.palfxdef)

		}
		return true
	})
	if e == nil {
		return false
	}
	// In this scenario the explod scale is constant in Mugen
	//if c.minus == -2 || c.minus == -4 {
	//	e.localscl = (320 / crun.localcoord)
	//} else {
	e.localscl = crun.localscl
	e.setStartParams(&e.palfxdef)
	e.setPos(crun)
	crun.insertExplodEx(i, rp)
	return false
}

func (sc explod) setInterpolation(c *Char, e *Explod,
	id byte, exp []BytecodeExp, pfd *PalFXDef) bool {
	switch id {
	case explod_interpolate_time:
		e.interpolate_time[0] = exp[0].evalI(c)
		if e.interpolate_time[0] < 0 {
			e.interpolate_time[0] = e.removetime
		}
		e.interpolate_time[1] = e.interpolate_time[0]
		if e.interpolate_time[0] > 0 {
			e.resetInterpolation(pfd)
			e.interpolate = true
			if e.ownpal {
				pfd.interpolate = true
				pfd.itime = e.interpolate_time[0]
			}
		}
	case explod_interpolate_animelem:
		e.interpolate_animelem[1] = exp[0].evalI(c)
		e.interpolate_animelem[0] = e.animelem
		e.interpolate_animelem[2] = e.interpolate_animelem[1]
	case explod_interpolate_pos:
		e.interpolate_pos[2] = exp[0].evalF(c)
		if len(exp) > 1 {
			e.interpolate_pos[3] = exp[1].evalF(c)
		}
	case explod_interpolate_scale:
		e.interpolate_scale[2] = exp[0].evalF(c)
		if len(exp) > 1 {
			e.interpolate_scale[3] = exp[1].evalF(c)
		}
	case explod_interpolate_alpha:
		e.interpolate_alpha[2] = exp[0].evalI(c)
		e.interpolate_alpha[3] = exp[1].evalI(c)
		e.interpolate_alpha[2] = Clamp(e.interpolate_alpha[2], 0, 255)
		e.interpolate_alpha[3] = Clamp(e.interpolate_alpha[3], 0, 255)
	case explod_interpolate_angle:
		e.interpolate_angle[3] = exp[0].evalF(c)
		if len(exp) > 1 {
			e.interpolate_angle[4] = exp[1].evalF(c)
		}
		if len(exp) > 2 {
			e.interpolate_angle[5] = exp[2].evalF(c)
		}
	case explod_interpolate_focallength:
		e.interpolate_fLength[1] = exp[0].evalF(c)
	case explod_interpolate_pfx_mul:
		pfd.imul[0] = exp[0].evalI(c)
		if len(exp) > 1 {
			pfd.imul[1] = exp[1].evalI(c)
		}
		if len(exp) > 2 {
			pfd.imul[2] = exp[2].evalI(c)
		}
	case explod_interpolate_pfx_add:
		pfd.iadd[0] = exp[0].evalI(c)
		if len(exp) > 1 {
			pfd.iadd[1] = exp[1].evalI(c)
		}
		if len(exp) > 2 {
			pfd.iadd[2] = exp[2].evalI(c)
		}
	case explod_interpolate_pfx_color:
		pfd.icolor[0] = exp[0].evalF(c) / 256
	case explod_interpolate_pfx_hue:
		pfd.ihue[0] = exp[0].evalF(c) / 256
	default:
	}
	return true
}

type modifyExplod explod

func (sc modifyExplod) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	eid := int32(-1)
	var expls []*Explod
	rp := [...]int32{-1, 0}
	remap := false
	var f, vf float32 = 1, 1
	sp, pos, vel, accel := Space_none, [2]float32{0, 0}, [2]float32{0, 0}, [2]float32{0, 0}
	ptexists := false
	eachExpl := func(f func(e *Explod)) {
		for _, e := range expls {
			f(e)
		}
	}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case explod_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
			} else {
				return false
			}
		case explod_remappal:
			rp[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				rp[1] = exp[1].evalI(c)
			}
			remap = true
		case explod_id:
			eid = exp[0].evalI(c)
		case explod_postypeExists:
			ptexists = true
		default:
			if len(expls) == 0 {
				expls = crun.getExplods(eid)
				if len(expls) == 0 {
					return false
				}
				eachExpl(func(e *Explod) {
					if e.ownpal && remap {
						crun.remapPal(e.palfx, [...]int32{1, 1}, rp)
					}
				})
			}
			switch id {
			case explod_facing:
				if exp[0].evalI(c) < 0 {
					f = -1
				}
				if (c.stWgi().ikemenver[0] != 0 || c.stWgi().ikemenver[1] != 0) && !ptexists {
					eachExpl(func(e *Explod) {
						e.relativef = f
					})
				}
			case explod_vfacing:
				if exp[0].evalI(c) < 0 {
					vf = -1
				}
				if (c.stWgi().ikemenver[0] != 0 || c.stWgi().ikemenver[1] != 0) && !ptexists {
					eachExpl(func(e *Explod) {
						e.vfacing = vf
					})
				}
			case explod_pos:
				pos[0] = exp[0].evalF(c) * lclscround
				if (c.stWgi().ikemenver[0] != 0 || c.stWgi().ikemenver[1] != 0) && !ptexists {
					eachExpl(func(e *Explod) { e.relativePos[0] = pos[0] })
				}
				if len(exp) > 1 {
					pos[1] = exp[1].evalF(c) * lclscround
					if (c.stWgi().ikemenver[0] != 0 || c.stWgi().ikemenver[1] != 0) && !ptexists {
						eachExpl(func(e *Explod) { e.relativePos[1] = pos[1] })
					}
				}
			case explod_random:
				rndx := (exp[0].evalF(c) / 2) * lclscround
				rndx = RandF(-rndx, rndx)
				pos[0] += rndx
				if (c.stWgi().ikemenver[0] != 0 || c.stWgi().ikemenver[1] != 0) && !ptexists {
					eachExpl(func(e *Explod) { e.relativePos[0] += rndx })
				}
				if len(exp) > 1 {
					rndy := (exp[1].evalF(c) / 2) * lclscround
					rndy = RandF(-rndy, rndy)
					pos[1] += rndy
					if (c.stWgi().ikemenver[0] != 0 || c.stWgi().ikemenver[1] != 0) && !ptexists {
						eachExpl(func(e *Explod) { e.relativePos[1] += rndy })
					}
				}
			case explod_velocity:
				vel[0] = exp[0].evalF(c) * lclscround
				if (c.stWgi().ikemenver[0] != 0 || c.stWgi().ikemenver[1] != 0) && !ptexists {
					eachExpl(func(e *Explod) { e.velocity[0] = vel[0] })
				}
				if len(exp) > 1 {
					vel[1] = exp[1].evalF(c) * lclscround
					if (c.stWgi().ikemenver[0] != 0 || c.stWgi().ikemenver[1] != 0) && !ptexists {
						eachExpl(func(e *Explod) { e.velocity[1] = vel[1] })
					}
				}
			case explod_accel:
				accel[0] = exp[0].evalF(c) * lclscround
				if (c.stWgi().ikemenver[0] != 0 || c.stWgi().ikemenver[1] != 0) && !ptexists {
					eachExpl(func(e *Explod) { e.accel[0] = accel[0] })
				}
				if len(exp) > 1 {
					accel[1] = exp[1].evalF(c) * lclscround
					if (c.stWgi().ikemenver[0] != 0 || c.stWgi().ikemenver[1] != 0) && !ptexists {
						eachExpl(func(e *Explod) { e.accel[1] = accel[1] })
					}
				}
			case explod_space:
				sp = Space(exp[0].evalI(c))
				if (c.stWgi().ikemenver[0] != 0 || c.stWgi().ikemenver[1] != 0) && !ptexists {
					eachExpl(func(e *Explod) { e.space = sp })
				}
			case explod_postype:
				pt := PosType(exp[0].evalI(c))
				eachExpl(func(e *Explod) {
					// Reset explod
					e.reset()
					// Set declared values
					e.postype = pt
					e.relativef, e.vfacing = f, vf
					e.relativePos, e.velocity, e.accel = pos, vel, accel
					if sp != Space_none {
						e.space = sp
					}
					// Finish pos configuration
					e.setPos(crun)
				})
			case explod_scale:
				x := exp[0].evalF(c)
				eachExpl(func(e *Explod) { e.scale[0] = x })
				if len(exp) > 1 {
					y := exp[1].evalF(c)
					eachExpl(func(e *Explod) { e.scale[1] = y })
				}
			case explod_bindtime:
				t := exp[0].evalI(c)
				eachExpl(func(e *Explod) {
					e.bindtime = t
					//Bindtime fix(update bindtime according to current explod time)
					if (crun.stWgi().ikemenver[0] > 0 || crun.stWgi().ikemenver[1] > 0) && t > 0 {
						e.bindtime = e.time + t
					}
					e.setX(e.pos[0])
					e.setY(e.pos[1])
				})
			case explod_removetime:
				t := exp[0].evalI(c)
				eachExpl(func(e *Explod) {
					e.removetime = t
					//Removetime fix(update removetime according to current explod time)
					if (crun.stWgi().ikemenver[0] > 0 || crun.stWgi().ikemenver[1] > 0) && t > 0 {
						e.removetime = e.time + t
					}
				})
			case explod_supermove:
				if exp[0].evalB(c) {
					eachExpl(func(e *Explod) { e.supermovetime = -1 })
				} else {
					eachExpl(func(e *Explod) { e.supermovetime = 0 })
				}
			case explod_supermovetime:
				t := exp[0].evalI(c)
				eachExpl(func(e *Explod) {
					e.supermovetime = t
					//Supermovetime fix(update supermovetime according to current explod time)
					if (crun.stWgi().ikemenver[0] > 0 || crun.stWgi().ikemenver[1] > 0) && t > 0 {
						e.supermovetime = e.time + t
					}
				})
			case explod_pausemovetime:
				t := exp[0].evalI(c)
				eachExpl(func(e *Explod) {
					e.pausemovetime = t
					//Pausemovetime fix(update pausemovetime according to current explod time)
					if (crun.stWgi().ikemenver[0] > 0 || crun.stWgi().ikemenver[1] > 0) && t > 0 {
						e.pausemovetime = e.time + t
					}
				})
			case explod_sprpriority:
				t := exp[0].evalI(c)
				eachExpl(func(e *Explod) { e.sprpriority = t })
			case explod_ontop:
				t := exp[0].evalB(c)
				eachExpl(func(e *Explod) {
					e.ontop = t
					if e.ontop && e.under {
						e.under = false
					}
				})
			case explod_strictontop:
				eachExpl(func(e *Explod) {
					if e.ontop {
						e.sprpriority = 0
					}
				})
			case explod_under:
				t := exp[0].evalB(c)
				eachExpl(func(e *Explod) {
					e.under = t
					if e.under && e.ontop {
						e.ontop = false
					}
				})
			case explod_shadow:
				r := exp[0].evalI(c)
				eachExpl(func(e *Explod) { e.shadow[0] = r })
				if len(exp) > 1 {
					g := exp[1].evalI(c)
					eachExpl(func(e *Explod) { e.shadow[1] = g })
					if len(exp) > 2 {
						b := exp[2].evalI(c)
						eachExpl(func(e *Explod) { e.shadow[2] = b })
					}
				}
			case explod_removeongethit:
				t := exp[0].evalB(c)
				eachExpl(func(e *Explod) { e.removeongethit = t })
			case explod_removeonchangestate:
				t := exp[0].evalB(c)
				eachExpl(func(e *Explod) { e.removeonchangestate = t })
			case explod_trans:
				s, d := exp[0].evalI(c), exp[1].evalI(c)
				blendmode := 0
				if len(exp) >= 3 {
					s, d = Clamp(s, 0, 255), Clamp(d, 0, 255)
					//if len(exp) >= 4 {
					//	d = ^d
					//} else if s == 1 && d == 255 {

					//Add
					blendmode = 1
					//Sub
					if s == 1 && d == 255 {
						blendmode = 2
					} else if s == -1 && d == 0 {
						blendmode = 0
					}

					if s == 1 && d == 255 {
						s = 0
					}

				}
				eachExpl(func(e *Explod) {
					e.alpha = [...]int32{s, d}
					e.blendmode = int32(blendmode)
				})
			case explod_anim:
				if c.stWgi().ikemenver[0] != 0 || c.stWgi().ikemenver[1] != 0 {
					anim := crun.getAnim(exp[1].evalI(c), string(*(*[]byte)(unsafe.Pointer(&exp[0]))), true)
					eachExpl(func(e *Explod) { e.anim = anim })
				}
			case explod_animelem:
				animelem := exp[0].evalI(c)
				eachExpl(func(e *Explod) {
					e.interpolate_animelem[1] = -1
					e.animelem = animelem
					e.anim.Action()
					e.setAnimElem()
				})
			case explod_animfreeze:
				animfreeze := exp[0].evalB(c)
				eachExpl(func(e *Explod) { e.animfreeze = animfreeze })
			case explod_angle:
				a := exp[0].evalF(c)
				eachExpl(func(e *Explod) { e.anglerot[0] = a })
			case explod_yangle:
				ya := exp[0].evalF(c)
				eachExpl(func(e *Explod) { e.anglerot[2] = ya })
			case explod_xangle:
				xa := exp[0].evalF(c)
				eachExpl(func(e *Explod) { e.anglerot[1] = xa })
			case explod_projection:
				eachExpl(func(e *Explod) { e.projection = Projection(exp[0].evalI(c)) })
			case explod_focallength:
				eachExpl(func(e *Explod) { e.fLength = exp[0].evalF(c) })
			case explod_window:
				eachExpl(func(e *Explod) {
					e.window = [4]float32{exp[0].evalF(c) * lclscround, exp[1].evalF(c) * lclscround, exp[2].evalF(c) * lclscround, exp[3].evalF(c) * lclscround}
				})
			case explod_ignorehitpause:
				if c.stWgi().ikemenver[0] != 0 || c.stWgi().ikemenver[1] != 0 {
					ihp := exp[0].evalB(c)
					eachExpl(func(e *Explod) { e.ignorehitpause = ihp })
				}
			case explod_bindid:
				bId := exp[0].evalI(c)
				if bId == -1 {
					bId = crun.id
				}
				eachExpl(func(e *Explod) { e.setBind(bId) })
			case explod_interpolation:
				if c.stWgi().ikemenver[0] != 0 || c.stWgi().ikemenver[1] != 0 {
					interpolation := exp[0].evalB(c)
					eachExpl(func(e *Explod) {
						if e.interpolate != interpolation && e.interpolate_time[0] > 0 {
							e.interpolate_animelem[0] = e.start_animelem
							e.interpolate_animelem[1] = e.interpolate_animelem[2]
							if e.ownpal {
								pfd := e.palfx
								pfd.interpolate = interpolation
								pfd.itime = e.interpolate_time[0]
							}
							e.interpolate_time[1] = e.interpolate_time[0]
							e.interpolate = interpolation
						}
					})
				}
			default:
				eachExpl(func(e *Explod) {
					if e.ownpal {
						palFX(sc).runSub(c, &e.palfx.PalFXDef, id, exp)
					}
				})
			}
		}
		return true
	})
	return false
}

type gameMakeAnim StateControllerBase

const (
	gameMakeAnim_pos byte = iota
	gameMakeAnim_random
	gameMakeAnim_under
	gameMakeAnim_anim
	gameMakeAnim_redirectid
)

func (sc gameMakeAnim) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	var e *Explod
	var i int
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if e == nil {
			if id == gameMakeAnim_redirectid {
				if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
					crun = rid
					lclscround = c.localscl / crun.localscl
					e, i = crun.newExplod()
					if e == nil {
						return false
					}
					e.id = 0
				} else {
					return false
				}
			} else {
				e, i = crun.newExplod()
				if e == nil {
					return false
				}
				e.id = 0
			}
			e.ontop, e.sprpriority, e.ownpal = true, math.MinInt32, true
		}
		switch id {
		case gameMakeAnim_pos:
			e.relativePos[0] = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				e.relativePos[1] = exp[1].evalF(c) * lclscround
			}
		case gameMakeAnim_random:
			rndx := (exp[0].evalF(c) / 2) * lclscround
			e.relativePos[0] += RandF(-rndx, rndx)
			if len(exp) > 1 {
				rndy := (exp[1].evalF(c) / 2) * lclscround
				e.relativePos[1] += RandF(-rndy, rndy)
			}
		case gameMakeAnim_under:
			e.ontop = !exp[0].evalB(c)
		case gameMakeAnim_anim:
			e.anim = crun.getAnim(exp[1].evalI(c), string(*(*[]byte)(unsafe.Pointer(&exp[0]))), true)
		}
		return true
	})
	if e == nil {
		return false
	}
	e.relativePos[0] -= float32(crun.size.draw.offset[0])
	e.relativePos[1] -= float32(crun.size.draw.offset[1])
	e.setPos(crun)
	crun.insertExplod(i)
	return false
}

type afterImage palFX

const (
	afterImage_trans = iota + palFX_last + 1
	afterImage_time
	afterImage_length
	afterImage_timegap
	afterImage_framegap
	afterImage_palcolor
	afterImage_palhue
	afterImage_palinvertall
	afterImage_palinvertblend
	afterImage_palbright
	afterImage_palcontrast
	afterImage_palpostbright
	afterImage_paladd
	afterImage_palmul
	afterImage_ignorehitpause
	afterImage_last = iota + palFX_last + 1 - 1
	afterImage_redirectid
)

func (sc afterImage) runSub(c *Char, ai *AfterImage,
	id byte, exp []BytecodeExp) {
	switch id {
	case afterImage_trans:
		ai.alpha[0] = exp[0].evalI(c)
		ai.alpha[1] = exp[1].evalI(c)
		if len(exp) >= 3 {
			ai.alpha[0] = Clamp(ai.alpha[0], 0, 255)
			ai.alpha[1] = Clamp(ai.alpha[1], 0, 255)
			//if len(exp) >= 4 {
			//	ai.alpha[1] = ^ai.alpha[1]
			//} else if ai.alpha[0] == 1 && ai.alpha[1] == 255 {
			if ai.alpha[0] == 1 && ai.alpha[1] == 255 {
				ai.alpha[0] = 0
			}
		}
	case afterImage_time:
		ai.time = exp[0].evalI(c)
	case afterImage_length:
		ai.length = exp[0].evalI(c)
	case afterImage_timegap:
		ai.timegap = Max(1, exp[0].evalI(c))
	case afterImage_framegap:
		ai.framegap = exp[0].evalI(c)
	case afterImage_palcolor:
		ai.setPalColor(exp[0].evalI(c))
	case afterImage_palhue:
		ai.setPalHueShift(exp[0].evalI(c))
	case afterImage_palinvertall:
		ai.setPalInvertall(exp[0].evalB(c))
	case afterImage_palinvertblend:
		ai.setPalInvertblend(exp[0].evalI(c))
	case afterImage_palbright:
		ai.setPalBrightR(exp[0].evalI(c))
		if len(exp) > 1 {
			ai.setPalBrightG(exp[1].evalI(c))
			if len(exp) > 2 {
				ai.setPalBrightB(exp[2].evalI(c))
			}
		}
	case afterImage_palcontrast:
		ai.setPalContrastR(exp[0].evalI(c))
		if len(exp) > 1 {
			ai.setPalContrastG(exp[1].evalI(c))
			if len(exp) > 2 {
				ai.setPalContrastB(exp[2].evalI(c))
			}
		}
	case afterImage_palpostbright:
		ai.postbright[0] = exp[0].evalI(c)
		if len(exp) > 1 {
			ai.postbright[1] = exp[1].evalI(c)
			if len(exp) > 2 {
				ai.postbright[2] = exp[2].evalI(c)
			}
		}
	case afterImage_paladd:
		ai.add[0] = exp[0].evalI(c)
		if len(exp) > 1 {
			ai.add[1] = exp[1].evalI(c)
			if len(exp) > 2 {
				ai.add[2] = exp[2].evalI(c)
			}
		}
	case afterImage_palmul:
		ai.mul[0] = exp[0].evalF(c)
		if len(exp) > 1 {
			ai.mul[1] = exp[1].evalF(c)
			if len(exp) > 2 {
				ai.mul[2] = exp[2].evalF(c)
			}
		}
	case afterImage_ignorehitpause:
		ai.ignorehitpause = exp[0].evalB(c)
	}
}
func (sc afterImage) Run(c *Char, _ []int32) bool {
	crun := c
	doOnce := false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if id == afterImage_redirectid {
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		if !doOnce {
			crun.aimg.clear()
			//Mugen 1.1 behavior if invertblend param is omitted(Only if char mugenversion = 1.1)
			if c.stWgi().mugenver[0] == 1 && c.stWgi().mugenver[1] == 1 && c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
				crun.aimg.palfx[0].invertblend = -2
			}
			crun.aimg.time = 1
			doOnce = true
		}
		sc.runSub(c, &crun.aimg, id, exp)
		return true
	})
	crun.aimg.setupPalFX()
	return false
}

type afterImageTime StateControllerBase

const (
	afterImageTime_time byte = iota
	afterImageTime_redirectid
)

func (sc afterImageTime) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if id == afterImageTime_redirectid {
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		if crun.aimg.timegap <= 0 {
			return false
		}
		switch id {
		case afterImageTime_time:
			time := exp[0].evalI(c)
			if time == 1 {
				time = 0
			}
			crun.aimg.time = time
			crun.aimg.timecount = 0
		}
		return true
	})
	return false
}

type hitDef afterImage

const (
	hitDef_attr = iota + afterImage_last + 1
	hitDef_guardflag
	hitDef_hitflag
	hitDef_ground_type
	hitDef_air_type
	hitDef_animtype
	hitDef_air_animtype
	hitDef_fall_animtype
	hitDef_affectteam
	hitDef_teamside
	hitDef_id
	hitDef_chainid
	hitDef_nochainid
	hitDef_kill
	hitDef_guard_kill
	hitDef_fall_kill
	hitDef_hitonce
	hitDef_air_juggle
	hitDef_getpower
	hitDef_damage
	hitDef_givepower
	hitDef_numhits
	hitDef_hitsound
	hitDef_hitsound_channel
	hitDef_guardsound
	hitDef_guardsound_channel
	hitDef_priority
	hitDef_p1stateno
	hitDef_p2stateno
	hitDef_p2getp1state
	hitDef_p1sprpriority
	hitDef_p2sprpriority
	hitDef_forcestand
	hitDef_forcecrouch
	hitDef_forcenofall
	hitDef_fall_damage
	hitDef_fall_xvelocity
	hitDef_fall_yvelocity
	hitDef_fall_recover
	hitDef_fall_recovertime
	hitDef_sparkno
	hitDef_sparkangle
	hitDef_guard_sparkno
	hitDef_guard_sparkangle
	hitDef_sparkxy
	hitDef_down_hittime
	hitDef_p1facing
	hitDef_p1getp2facing
	hitDef_mindist
	hitDef_maxdist
	hitDef_snap
	hitDef_p2facing
	hitDef_air_hittime
	hitDef_fall
	hitDef_air_fall
	hitDef_air_cornerpush_veloff
	hitDef_down_bounce
	hitDef_down_velocity
	hitDef_down_cornerpush_veloff
	hitDef_ground_hittime
	hitDef_guard_hittime
	hitDef_guard_dist
	hitDef_guard_dist_back
	hitDef_pausetime
	hitDef_guard_pausetime
	hitDef_air_velocity
	hitDef_airguard_velocity
	hitDef_ground_slidetime
	hitDef_guard_slidetime
	hitDef_guard_ctrltime
	hitDef_airguard_ctrltime
	hitDef_ground_velocity_x
	hitDef_ground_velocity_y
	hitDef_ground_velocity
	hitDef_guard_velocity
	hitDef_ground_cornerpush_veloff
	hitDef_guard_cornerpush_veloff
	hitDef_airguard_cornerpush_veloff
	hitDef_yaccel
	hitDef_envshake_time
	hitDef_envshake_ampl
	hitDef_envshake_phase
	hitDef_envshake_freq
	hitDef_envshake_mul
	hitDef_fall_envshake_time
	hitDef_fall_envshake_ampl
	hitDef_fall_envshake_phase
	hitDef_fall_envshake_freq
	hitDef_fall_envshake_mul
	hitDef_dizzypoints
	hitDef_guardpoints
	hitDef_redlife
	hitDef_score
	hitDef_last = iota + afterImage_last + 1 - 1
	hitDef_redirectid
)

func (sc hitDef) runSub(c *Char, hd *HitDef, id byte, exp []BytecodeExp) bool {
	switch id {
	case hitDef_attr:
		hd.attr = exp[0].evalI(c)
	case hitDef_guardflag:
		hd.guardflag = exp[0].evalI(c)
	case hitDef_hitflag:
		hd.hitflag = exp[0].evalI(c)
	case hitDef_ground_type:
		hd.ground_type = HitType(exp[0].evalI(c))
	case hitDef_air_type:
		hd.air_type = HitType(exp[0].evalI(c))
	case hitDef_animtype:
		hd.animtype = Reaction(exp[0].evalI(c))
	case hitDef_air_animtype:
		hd.air_animtype = Reaction(exp[0].evalI(c))
	case hitDef_fall_animtype:
		hd.fall.animtype = Reaction(exp[0].evalI(c))
	case hitDef_affectteam:
		hd.affectteam = exp[0].evalI(c)
	case hitDef_teamside:
		n := exp[0].evalI(c)
		if n > 2 {
			hd.teamside = 2
		} else if n < 0 {
			hd.teamside = 0
		} else {
			hd.teamside = int(n)
		}
	case hitDef_id:
		hd.id = Max(0, exp[0].evalI(c))
	case hitDef_chainid:
		hd.chainid = exp[0].evalI(c)
	case hitDef_nochainid:
		hd.nochainid[0] = exp[0].evalI(c)
		if len(exp) > 1 {
			hd.nochainid[1] = exp[1].evalI(c)
		}
	case hitDef_kill:
		hd.kill = exp[0].evalB(c)
	case hitDef_guard_kill:
		hd.guard_kill = exp[0].evalB(c)
	case hitDef_fall_kill:
		hd.fall.kill = exp[0].evalB(c)
	case hitDef_hitonce:
		hd.hitonce = Btoi(exp[0].evalB(c))
	case hitDef_air_juggle:
		hd.air_juggle = exp[0].evalI(c)
	case hitDef_getpower:
		hd.hitgetpower = Max(IErr+1, exp[0].evalI(c))
		if len(exp) > 1 {
			hd.guardgetpower = Max(IErr+1, exp[1].evalI(c))
		}
	case hitDef_damage:
		hd.hitdamage = exp[0].evalI(c)
		if len(exp) > 1 {
			hd.guarddamage = exp[1].evalI(c)
		}
	case hitDef_givepower:
		hd.hitgivepower = Max(IErr+1, exp[0].evalI(c))
		if len(exp) > 1 {
			hd.guardgivepower = Max(IErr+1, exp[1].evalI(c))
		}
	case hitDef_numhits:
		hd.numhits = exp[0].evalI(c)
	case hitDef_hitsound:
		hd.hitsound_ffx = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		hd.hitsound[0] = exp[1].evalI(c)
		if len(exp) > 2 {
			hd.hitsound[1] = exp[2].evalI(c)
		}
	case hitDef_hitsound_channel:
		hd.hitsound_channel = exp[0].evalI(c)
	case hitDef_guardsound:
		hd.guardsound_ffx = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		hd.guardsound[0] = exp[1].evalI(c)
		if len(exp) > 2 {
			hd.guardsound[1] = exp[2].evalI(c)
		}
	case hitDef_guardsound_channel:
		hd.guardsound_channel = exp[0].evalI(c)
	case hitDef_priority:
		hd.priority = exp[0].evalI(c)
		hd.bothhittype = AiuchiType(exp[1].evalI(c))
	case hitDef_p1stateno:
		hd.p1stateno = exp[0].evalI(c)
	case hitDef_p2stateno:
		hd.p2stateno = exp[0].evalI(c)
		hd.p2getp1state = true
	case hitDef_p2getp1state:
		hd.p2getp1state = exp[0].evalB(c)
	case hitDef_p1sprpriority:
		hd.p1sprpriority = exp[0].evalI(c)
	case hitDef_p2sprpriority:
		hd.p2sprpriority = exp[0].evalI(c)
	case hitDef_forcestand:
		hd.forcestand = Btoi(exp[0].evalB(c))
	case hitDef_forcecrouch:
		hd.forcecrouch = Btoi(exp[0].evalB(c))
	case hitDef_forcenofall:
		hd.forcenofall = exp[0].evalB(c)
	case hitDef_fall_damage:
		hd.fall.damage = exp[0].evalI(c)
	case hitDef_fall_xvelocity:
		hd.fall.xvelocity = exp[0].evalF(c)
	case hitDef_fall_yvelocity:
		hd.fall.yvelocity = exp[0].evalF(c)
	case hitDef_fall_recover:
		hd.fall.recover = exp[0].evalB(c)
	case hitDef_fall_recovertime:
		hd.fall.recovertime = exp[0].evalI(c)
	case hitDef_sparkno:
		hd.sparkno_ffx = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		hd.sparkno = exp[1].evalI(c)
	case hitDef_sparkangle:
		hd.sparkangle = exp[0].evalF(c)
	case hitDef_guard_sparkno:
		hd.guard_sparkno_ffx = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		hd.guard_sparkno = exp[1].evalI(c)
	case hitDef_guard_sparkangle:
		hd.guard_sparkangle = exp[0].evalF(c)
	case hitDef_sparkxy:
		hd.sparkxy[0] = exp[0].evalF(c)
		if len(exp) > 1 {
			hd.sparkxy[1] = exp[1].evalF(c)
		}
	case hitDef_down_hittime:
		hd.down_hittime = exp[0].evalI(c)
	case hitDef_p1facing:
		hd.p1facing = exp[0].evalI(c)
	case hitDef_p1getp2facing:
		hd.p1getp2facing = exp[0].evalI(c)
	case hitDef_mindist:
		hd.mindist[0] = exp[0].evalF(c)
		if len(exp) > 1 {
			hd.mindist[1] = exp[1].evalF(c)
			if len(exp) > 2 {
				exp[2].run(c)
			}
		}
	case hitDef_maxdist:
		hd.maxdist[0] = exp[0].evalF(c)
		if len(exp) > 1 {
			hd.maxdist[1] = exp[1].evalF(c)
			if len(exp) > 2 {
				exp[2].run(c)
			}
		}
	case hitDef_snap:
		hd.snap[0] = exp[0].evalF(c)
		if len(exp) > 1 {
			hd.snap[1] = exp[1].evalF(c)
			if len(exp) > 2 {
				exp[2].run(c)
				if len(exp) > 3 {
					hd.snapt = exp[3].evalI(c)
				}
			}
		}
	case hitDef_p2facing:
		hd.p2facing = exp[0].evalI(c)
	case hitDef_air_hittime:
		hd.air_hittime = exp[0].evalI(c)
	case hitDef_fall:
		hd.ground_fall = exp[0].evalB(c)
		hd.air_fall = hd.ground_fall
	case hitDef_air_fall:
		hd.air_fall = exp[0].evalB(c)
	case hitDef_air_cornerpush_veloff:
		hd.air_cornerpush_veloff = exp[0].evalF(c)
	case hitDef_down_bounce:
		hd.down_bounce = exp[0].evalB(c)
	case hitDef_down_velocity:
		hd.down_velocity[0] = exp[0].evalF(c)
		if len(exp) > 1 {
			hd.down_velocity[1] = exp[1].evalF(c)
		}
	case hitDef_down_cornerpush_veloff:
		hd.down_cornerpush_veloff = exp[0].evalF(c)
	case hitDef_ground_hittime:
		hd.ground_hittime = exp[0].evalI(c)
		hd.guard_hittime = hd.ground_hittime
	case hitDef_guard_hittime:
		hd.guard_hittime = exp[0].evalI(c)
	case hitDef_guard_dist:
		hd.guard_dist[0] = exp[0].evalI(c)
	case hitDef_guard_dist_back:
		hd.guard_dist[1] = exp[0].evalI(c)
	case hitDef_pausetime:
		hd.pausetime = exp[0].evalI(c)
		hd.guard_pausetime = hd.pausetime
		if len(exp) > 1 {
			hd.shaketime = exp[1].evalI(c)
			hd.guard_shaketime = hd.shaketime
		}
	case hitDef_guard_pausetime:
		hd.guard_pausetime = exp[0].evalI(c)
		if len(exp) > 1 {
			hd.guard_shaketime = exp[1].evalI(c)
		}
	case hitDef_air_velocity:
		hd.air_velocity[0] = exp[0].evalF(c)
		if len(exp) > 1 {
			hd.air_velocity[1] = exp[1].evalF(c)
		}
	case hitDef_airguard_velocity:
		hd.airguard_velocity[0] = exp[0].evalF(c)
		if len(exp) > 1 {
			hd.airguard_velocity[1] = exp[1].evalF(c)
		}
	case hitDef_ground_slidetime:
		hd.ground_slidetime = exp[0].evalI(c)
		hd.guard_slidetime = hd.ground_slidetime
		hd.guard_ctrltime = hd.ground_slidetime
		hd.airguard_ctrltime = hd.ground_slidetime
	case hitDef_guard_slidetime:
		hd.guard_slidetime = exp[0].evalI(c)
		hd.guard_ctrltime = hd.guard_slidetime
		hd.airguard_ctrltime = hd.guard_slidetime
	case hitDef_guard_ctrltime:
		hd.guard_ctrltime = exp[0].evalI(c)
		hd.airguard_ctrltime = hd.guard_ctrltime
	case hitDef_airguard_ctrltime:
		hd.airguard_ctrltime = exp[0].evalI(c)
	case hitDef_ground_velocity_x:
		hd.ground_velocity[0] = exp[0].evalF(c)
	case hitDef_ground_velocity_y:
		hd.ground_velocity[1] = exp[0].evalF(c)
	case hitDef_guard_velocity:
		hd.guard_velocity = exp[0].evalF(c)
	case hitDef_ground_cornerpush_veloff:
		hd.ground_cornerpush_veloff = exp[0].evalF(c)
	case hitDef_guard_cornerpush_veloff:
		hd.guard_cornerpush_veloff = exp[0].evalF(c)
	case hitDef_airguard_cornerpush_veloff:
		hd.airguard_cornerpush_veloff = exp[0].evalF(c)
	case hitDef_yaccel:
		hd.yaccel = exp[0].evalF(c)
	case hitDef_envshake_time:
		hd.envshake_time = exp[0].evalI(c)
	case hitDef_envshake_ampl:
		hd.envshake_ampl = exp[0].evalI(c)
	case hitDef_envshake_phase:
		hd.envshake_phase = exp[0].evalF(c)
	case hitDef_envshake_freq:
		hd.envshake_freq = MaxF(0, exp[0].evalF(c))
	case hitDef_envshake_mul:
		hd.envshake_mul = exp[0].evalF(c)
	case hitDef_fall_envshake_time:
		hd.fall.envshake_time = exp[0].evalI(c)
	case hitDef_fall_envshake_ampl:
		hd.fall.envshake_ampl = exp[0].evalI(c)
	case hitDef_fall_envshake_phase:
		hd.fall.envshake_phase = exp[0].evalF(c)
	case hitDef_fall_envshake_freq:
		hd.fall.envshake_freq = MaxF(0, exp[0].evalF(c))
	case hitDef_fall_envshake_mul:
		hd.fall.envshake_mul = exp[0].evalF(c)
	case hitDef_dizzypoints:
		hd.dizzypoints = Max(IErr+1, exp[0].evalI(c))
	case hitDef_guardpoints:
		hd.guardpoints = Max(IErr+1, exp[0].evalI(c))
	case hitDef_redlife:
		hd.hitredlife = Max(IErr+1, exp[0].evalI(c))
		if len(exp) > 1 {
			hd.guardredlife = exp[1].evalI(c)
		}
	case hitDef_score:
		hd.score[0] = exp[0].evalF(c)
		if len(exp) > 1 {
			hd.score[1] = exp[1].evalF(c)
		}
	default:
		if !palFX(sc).runSub(c, &hd.palfx, id, exp) {
			return false
		}
	}
	return true
}
func (sc hitDef) Run(c *Char, _ []int32) bool {
	crun := c
	crun.hitdef.clear()
	crun.hitdef.playerNo = sys.workingState.playerNo
	crun.hitdef.sparkno = c.gi().data.sparkno
	crun.hitdef.guard_sparkno = c.gi().data.guard.sparkno
	crun.hitdef.hitsound_channel = c.gi().data.hitsound_channel
	crun.hitdef.guardsound_channel = c.gi().data.guardsound_channel
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if id == hitDef_redirectid {
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				crun.hitdef.clear()
				crun.hitdef.playerNo = sys.workingState.playerNo
				crun.hitdef.sparkno = c.gi().data.sparkno
				crun.hitdef.guard_sparkno = c.gi().data.guard.sparkno
			} else {
				return false
			}
		}
		//Mugen 1.1 behavior if invertblend param is omitted(Only if char mugenversion = 1.1)
		if c.stWgi().mugenver[0] == 1 && c.stWgi().mugenver[1] == 1 && c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
			crun.hitdef.palfx.invertblend = -2
		}
		sc.runSub(c, &crun.hitdef, id, exp)
		return true
	})
	//winmugenでHitdefのattrが投げ属性で自分側pausetimeが1以上の時、毎フレーム実行されなくなる
	//"In Winmugen, when the attr of Hitdef is set to 'Throw' and the pausetime
	// on the attacker's side is greater than 1, it no longer executes every frame."
	if crun.hitdef.attr&int32(AT_AT) != 0 && crun.moveContact() == 1 &&
		c.gi().mugenver[0] != 1 && crun.hitdef.pausetime > 0 {
		crun.hitdef.attr = 0
		return false
	}
	crun.setHitdefDefault(&crun.hitdef, false)
	return false
}

type reversalDef hitDef

const (
	reversalDef_reversal_attr = iota + hitDef_last + 1
	reversalDef_redirectid
)

func (sc reversalDef) Run(c *Char, _ []int32) bool {
	crun := c
	crun.hitdef.clear()
	crun.hitdef.playerNo = sys.workingState.playerNo
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case reversalDef_reversal_attr:
			crun.hitdef.reversal_attr = exp[0].evalI(c)
		case reversalDef_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				crun.hitdef.clear()
				crun.hitdef.playerNo = sys.workingState.playerNo
			} else {
				return false
			}
		default:
			hitDef(sc).runSub(c, &crun.hitdef, id, exp)
		}
		return true
	})
	crun.setHitdefDefault(&crun.hitdef, false)
	return false
}

type projectile hitDef

const (
	projectile_postype = iota + hitDef_last + 1
	projectile_projid
	projectile_projremove
	projectile_projremovetime
	projectile_projshadow
	projectile_projmisstime
	projectile_projhits
	projectile_projpriority
	projectile_projhitanim
	projectile_projremanim
	projectile_projcancelanim
	projectile_velocity
	projectile_velmul
	projectile_remvelocity
	projectile_accel
	projectile_projscale
	projectile_projangle
	projectile_projrescaleclsn
	projectile_offset
	projectile_projsprpriority
	projectile_projstagebound
	projectile_projedgebound
	projectile_projheightbound
	projectile_projanim
	projectile_supermovetime
	projectile_pausemovetime
	projectile_ownpal
	projectile_remappal
	// projectile_platform
	// projectile_platformwidth
	// projectile_platformheight
	// projectile_platformfence
	// projectile_platformangle
	projectile_redirectid
)

func (sc projectile) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	var p *Projectile
	pt := PT_P1
	var x, y float32 = 0, 0
	op := false
	rc := false
	rp := [...]int32{-1, 0}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if p == nil {
			if id == projectile_redirectid {
				if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
					crun = rid
					lclscround = c.localscl / crun.localscl
					p = crun.newProj()
					if p == nil {
						return false
					}
					p.hitdef.playerNo = sys.workingState.playerNo

				} else {
					return false
				}
			} else {
				p = crun.newProj()
				if p == nil {
					return false
				}
				p.hitdef.playerNo = sys.workingState.playerNo
			}
		}
		switch id {
		case projectile_postype:
			pt = PosType(exp[0].evalI(c))
		case projectile_projid:
			p.id = exp[0].evalI(c)
		case projectile_projremove:
			p.remove = exp[0].evalB(c)
		case projectile_projremovetime:
			p.removetime = exp[0].evalI(c)
		case projectile_projshadow:
			p.shadow[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				p.shadow[1] = exp[1].evalI(c)
				if len(exp) > 2 {
					p.shadow[2] = exp[2].evalI(c)
				}
			}
		case projectile_projmisstime:
			p.misstime = exp[0].evalI(c)
		case projectile_projhits:
			p.hits = exp[0].evalI(c)
		case projectile_projpriority:
			p.priority = exp[0].evalI(c)
			p.priorityPoints = p.priority
		case projectile_projhitanim:
			p.hitanim = exp[1].evalI(c)
			p.hitanim_ffx = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		case projectile_projremanim:
			p.remanim = Max(-1, exp[1].evalI(c))
			p.remanim_ffx = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		case projectile_projcancelanim:
			p.cancelanim = Max(-1, exp[1].evalI(c))
			p.cancelanim_ffx = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		case projectile_velocity:
			p.velocity[0] = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				p.velocity[1] = exp[1].evalF(c) * lclscround
			}
		case projectile_velmul:
			p.velmul[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				p.velmul[1] = exp[1].evalF(c)
			}
		case projectile_remvelocity:
			p.remvelocity[0] = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				p.remvelocity[1] = exp[1].evalF(c) * lclscround
			}
		case projectile_accel:
			p.accel[0] = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				p.accel[1] = exp[1].evalF(c) * lclscround
			}
		case projectile_projscale:
			p.scale[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				p.scale[1] = exp[1].evalF(c)
			}
		case projectile_projangle:
			p.angle = exp[0].evalF(c)
		case projectile_offset:
			x = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				y = exp[1].evalF(c) * lclscround
			}
		case projectile_projsprpriority:
			p.sprpriority = exp[0].evalI(c)
		case projectile_projstagebound:
			p.stagebound = int32(float32(exp[0].evalI(c)) * lclscround)
		case projectile_projedgebound:
			p.edgebound = int32(float32(exp[0].evalI(c)) * lclscround)
		case projectile_projheightbound:
			p.heightbound[0] = int32(float32(exp[0].evalI(c)) * lclscround)
			if len(exp) > 1 {
				p.heightbound[1] = int32(float32(exp[1].evalI(c)) * lclscround)
			}
		case projectile_projanim:
			p.anim = exp[1].evalI(c)
			p.anim_ffx = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		case projectile_supermovetime:
			p.supermovetime = exp[0].evalI(c)
			if p.supermovetime >= 0 {
				p.supermovetime = Max(p.supermovetime, p.supermovetime+1)
			}
		case projectile_pausemovetime:
			p.pausemovetime = exp[0].evalI(c)
			if p.pausemovetime >= 0 {
				p.pausemovetime = Max(p.pausemovetime, p.pausemovetime+1)
			}
		case projectile_ownpal:
			op = exp[0].evalB(c)
		case projectile_remappal:
			rp[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				rp[1] = exp[1].evalI(c)
			}
		case projectile_projrescaleclsn:
			rc = exp[0].evalB(c)
		// case projectile_platform:
		// 	p.platform = exp[0].evalB(c)
		// case projectile_platformwidth:
		// 	p.platformWidth[0] = exp[0].evalF(c) * lclscround
		// 	if len(exp) > 1 {
		// 		p.platformWidth[1] = exp[1].evalF(c) * lclscround
		// 	}
		// case projectile_platformheight:
		// 	p.platformHeight[0] = exp[0].evalF(c) * lclscround
		// 	if len(exp) > 1 {
		// 		p.platformHeight[1] = exp[1].evalF(c) * lclscround
		// 	}
		// case projectile_platformangle:
		// 	p.platformAngle = exp[0].evalF(c)
		// case projectile_platformfence:
		// 	p.platformFence = exp[0].evalB(c)
		default:
			if !hitDef(sc).runSub(c, &p.hitdef, id, exp) {
				afterImage(sc).runSub(c, &p.aimg, id, exp)
			}
		}
		return true
	})
	if p == nil {
		return false
	}
	crun.setHitdefDefault(&p.hitdef, true)
	if p.hitanim == -1 {
		p.hitanim_ffx = p.anim_ffx
	}
	if p.remanim == IErr {
		p.remanim = p.hitanim
		p.remanim_ffx = p.hitanim_ffx
	}
	if p.cancelanim == IErr {
		p.cancelanim = p.remanim
		p.cancelanim_ffx = p.remanim_ffx
	}
	if p.aimg.time != 0 {
		p.aimg.setupPalFX()
	}
	if crun.minus == -2 || crun.minus == -4 {
		p.localscl = (320 / crun.localcoord)
	} else {
		p.localscl = crun.localscl
	}
	crun.projInit(p, pt, x, y, op, rp[0], rp[1], rc)
	return false
}

type width StateControllerBase

const (
	width_edge byte = iota
	width_player
	width_value
	width_redirectid
)

func (sc width) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case width_edge:
			crun.setFEdge(exp[0].evalF(c) * lclscround)
			if len(exp) > 1 {
				crun.setBEdge(exp[1].evalF(c) * lclscround)
			}
		case width_player:
			crun.setFWidth(exp[0].evalF(c) * lclscround)
			if len(exp) > 1 {
				crun.setBWidth(exp[1].evalF(c) * lclscround)
			}
		case width_value:
			v1 := exp[0].evalF(c) * lclscround
			crun.setFEdge(v1)
			crun.setFWidth(v1)
			if len(exp) > 1 {
				v2 := exp[1].evalF(c) * lclscround
				crun.setBEdge(v2)
				crun.setBWidth(v2)
			}
		case width_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = (320 / c.localcoord) / (320 / crun.localcoord)
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type sprPriority StateControllerBase

const (
	sprPriority_value byte = iota
	sprPriority_redirectid
)

func (sc sprPriority) Run(c *Char, _ []int32) bool {
	crun := c
	v := int32(0) // Mugen uses 0 if no value is set at all
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case sprPriority_value:
			v = exp[0].evalI(c)
		case sprPriority_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.setSprPriority(v)
	return false
}

type varSet StateControllerBase

const (
	varSet_ byte = iota
	varSet_redirectid
)

func (sc varSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case varSet_:
			exp[0].run(crun)
		case varSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type turn StateControllerBase

const (
	turn_ byte = iota
	turn_redirectid
)

func (sc turn) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case turn_:
			crun.setFacing(-crun.facing)
		case turn_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type targetFacing StateControllerBase

const (
	targetFacing_id byte = iota
	targetFacing_value
	targetFacing_redirectid
)

func (sc targetFacing) Run(c *Char, _ []int32) bool {
	crun := c
	tar := crun.getTarget(-1)
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetFacing_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case targetFacing_value:
			if len(tar) == 0 {
				return false
			}
			crun.targetFacing(tar, exp[0].evalI(c))
		case targetFacing_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}

		}
		return true
	})
	if len(tar) == 0 {
		return false
	}
	return false
}

type targetBind StateControllerBase

const (
	targetBind_id byte = iota
	targetBind_time
	targetBind_pos
	targetBind_redirectid
)

func (sc targetBind) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	tar := crun.getTarget(-1)
	t := int32(1)
	var x, y float32 = 0, 0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetBind_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case targetBind_time:
			t = exp[0].evalI(c)
		case targetBind_pos:
			x = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				y = exp[1].evalF(c) * lclscround
			}
		case targetBind_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}

		}
		return true
	})
	if len(tar) == 0 {
		return false
	}
	crun.targetBind(tar, t, x, y)
	return false
}

type bindToTarget StateControllerBase

const (
	bindToTarget_id byte = iota
	bindToTarget_time
	bindToTarget_pos
	bindToTarget_redirectid
)

func (sc bindToTarget) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	tar := crun.getTarget(-1)
	t, x, y, hmf := int32(1), float32(0), float32(math.NaN()), HMF_F
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case bindToTarget_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case bindToTarget_time:
			t = exp[0].evalI(c)
		case bindToTarget_pos:
			x = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				y = exp[1].evalF(c) * lclscround
				if len(exp) > 2 {
					hmf = HMF(exp[2].evalI(c))
				}
			}
		case bindToTarget_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}
		}
		return true
	})
	if len(tar) == 0 {
		return false
	}
	crun.bindToTarget(tar, t, x, y, hmf)
	return false
}

type targetLifeAdd StateControllerBase

const (
	targetLifeAdd_id byte = iota
	targetLifeAdd_absolute
	targetLifeAdd_kill
	targetLifeAdd_dizzy
	targetLifeAdd_redlife
	targetLifeAdd_value
	targetLifeAdd_redirectid
)

func (sc targetLifeAdd) Run(c *Char, _ []int32) bool {
	crun := c
	tar, a, k, d, r := crun.getTarget(-1), false, true, true, true
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetLifeAdd_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case targetLifeAdd_absolute:
			a = exp[0].evalB(c)
		case targetLifeAdd_kill:
			k = exp[0].evalB(c)
		case targetLifeAdd_dizzy:
			d = exp[0].evalB(c)
		case targetLifeAdd_redlife:
			r = exp[0].evalB(c)
		case targetLifeAdd_value:
			if len(tar) == 0 {
				return false
			}
			crun.targetLifeAdd(tar, exp[0].evalI(c), k, a, d, r)
		case targetLifeAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type targetState StateControllerBase

const (
	targetState_id byte = iota
	targetState_value
	targetState_redirectid
)

func (sc targetState) Run(c *Char, _ []int32) bool {
	crun := c
	tar := crun.getTarget(-1)
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetState_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case targetState_value:
			if len(tar) == 0 {
				return false
			}
			crun.targetState(tar, exp[0].evalI(c))
		case targetState_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type targetVelSet StateControllerBase

const (
	targetVelSet_id byte = iota
	targetVelSet_x
	targetVelSet_y
	targetVelSet_redirectid
)

func (sc targetVelSet) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	tar := crun.getTarget(-1)
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetVelSet_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case targetVelSet_x:
			if len(tar) == 0 {
				return false
			}
			crun.targetVelSetX(tar, exp[0].evalF(c)*lclscround)
		case targetVelSet_y:
			if len(tar) == 0 {
				return false
			}
			crun.targetVelSetY(tar, exp[0].evalF(c)*lclscround)
		case targetVelSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type targetVelAdd StateControllerBase

const (
	targetVelAdd_id byte = iota
	targetVelAdd_x
	targetVelAdd_y
	targetVelAdd_redirectid
)

func (sc targetVelAdd) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	tar := crun.getTarget(-1)
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetVelAdd_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case targetVelAdd_x:
			if len(tar) == 0 {
				return false
			}
			crun.targetVelAddX(tar, exp[0].evalF(c)*lclscround)
		case targetVelAdd_y:
			if len(tar) == 0 {
				return false
			}
			crun.targetVelAddY(tar, exp[0].evalF(c)*lclscround)
		case targetVelAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type targetPowerAdd StateControllerBase

const (
	targetPowerAdd_id byte = iota
	targetPowerAdd_value
	targetPowerAdd_redirectid
)

func (sc targetPowerAdd) Run(c *Char, _ []int32) bool {
	crun := c
	tar := crun.getTarget(-1)
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetPowerAdd_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case targetPowerAdd_value:
			if len(tar) == 0 {
				return false
			}
			crun.targetPowerAdd(tar, exp[0].evalI(c))
		case targetPowerAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type targetDrop StateControllerBase

const (
	targetDrop_excludeid byte = iota
	targetDrop_keepone
	targetDrop_redirectid
)

func (sc targetDrop) Run(c *Char, _ []int32) bool {
	crun := c
	tar, eid, ko := crun.getTarget(-1), int32(-1), true
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetDrop_excludeid:
			eid = exp[0].evalI(c)
		case targetDrop_keepone:
			ko = exp[0].evalB(c)
		case targetDrop_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}
		}
		return true
	})
	if len(tar) == 0 {
		return false
	}
	crun.targetDrop(eid, -1, ko)
	return false
}

type lifeAdd StateControllerBase

const (
	lifeAdd_absolute byte = iota
	lifeAdd_kill
	lifeAdd_value
	lifeAdd_redirectid
)

func (sc lifeAdd) Run(c *Char, _ []int32) bool {
	a, k := false, true
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case lifeAdd_absolute:
			a = exp[0].evalB(c)
		case lifeAdd_kill:
			k = exp[0].evalB(c)
		case lifeAdd_value:
			v := exp[0].evalI(c)
			// Mugen forces absolute parameter when healing characters
			if v > 0 && c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
				a = true
			}
			crun.lifeAdd(float64(v), k, a)
			crun.ghv.kill = k // The kill GetHitVar must currently be set here because c.lifeAdd is also used internally
		case lifeAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type lifeSet StateControllerBase

const (
	lifeSet_value byte = iota
	lifeSet_redirectid
)

func (sc lifeSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case lifeSet_value:
			crun.lifeSet(exp[0].evalI(c))
		case lifeSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type powerAdd StateControllerBase

const (
	powerAdd_value byte = iota
	powerAdd_redirectid
)

func (sc powerAdd) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case powerAdd_value:
			crun.powerAdd(exp[0].evalI(c))
		case powerAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type powerSet StateControllerBase

const (
	powerSet_value byte = iota
	powerSet_redirectid
)

func (sc powerSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case powerSet_value:
			crun.powerSet(exp[0].evalI(c))
		case powerSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type hitVelSet StateControllerBase

const (
	hitVelSet_x byte = iota
	hitVelSet_y
	hitVelSet_redirectid
)

func (sc hitVelSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case hitVelSet_x:
			if exp[0].evalB(c) {
				crun.hitVelSetX()
			}
		case hitVelSet_y:
			if exp[0].evalB(c) {
				crun.hitVelSetY()
			}
		case hitVelSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type screenBound StateControllerBase

const (
	screenBound_value byte = iota
	screenBound_movecamera
	screenBound_stagebound
	screenBound_redirectid
)

func (sc screenBound) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case screenBound_value:
			if exp[0].evalB(c) {
				crun.setCSF(CSF_screenbound)
			} else {
				crun.unsetCSF(CSF_screenbound)
			}
		case screenBound_movecamera:
			if exp[0].evalB(c) {
				crun.setCSF(CSF_movecamera_x)
			} else {
				crun.unsetCSF(CSF_movecamera_x)
			}
			if len(exp) > 1 {
				if exp[1].evalB(c) {
					crun.setCSF(CSF_movecamera_y)
				} else {
					crun.unsetCSF(CSF_movecamera_y)
				}
			} else {
				crun.unsetCSF(CSF_movecamera_y)
			}
		case screenBound_stagebound:
			if exp[0].evalB(c) {
				crun.setCSF(CSF_stagebound)
			} else {
				crun.unsetCSF(CSF_stagebound)
			}
		case screenBound_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type posFreeze StateControllerBase

const (
	posFreeze_value byte = iota
	posFreeze_redirectid
)

func (sc posFreeze) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case posFreeze_value:
			if exp[0].evalB(c) {
				crun.setCSF(CSF_posfreeze)
			}
		case posFreeze_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type envShake StateControllerBase

const (
	envShake_time byte = iota
	envShake_ampl
	envShake_phase
	envShake_freq
	envShake_mul
)

func (sc envShake) Run(c *Char, _ []int32) bool {
	sys.envShake.clear()
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case envShake_time:
			sys.envShake.time = exp[0].evalI(c)
		case envShake_ampl:
			sys.envShake.ampl = float32(int32(float32(exp[0].evalI(c)) * c.localscl))
		case envShake_phase:
			sys.envShake.phase = MaxF(0, exp[0].evalF(c)*float32(math.Pi)/180) * c.localscl
		case envShake_freq:
			sys.envShake.freq = MaxF(0, exp[0].evalF(c)*float32(math.Pi)/180)
		case envShake_mul:
			sys.envShake.mul = exp[0].evalF(c)
		}
		return true
	})
	sys.envShake.setDefPhase()
	return false
}

type hitOverride StateControllerBase

const (
	hitOverride_attr byte = iota
	hitOverride_slot
	hitOverride_stateno
	hitOverride_time
	hitOverride_forceair
	hitOverride_keepstate
	hitOverride_redirectid
)

func (sc hitOverride) Run(c *Char, _ []int32) bool {
	crun := c
	var a, s, st, t int32 = 0, 0, -1, 1
	f := false
	ks := false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case hitOverride_attr:
			a = exp[0].evalI(c)
		case hitOverride_slot:
			s = Max(0, exp[0].evalI(c))
			if s > 7 {
				s = 0
			}
		case hitOverride_stateno:
			st = exp[0].evalI(c)
		case hitOverride_time:
			t = exp[0].evalI(c)
			if t < -1 || t == 0 {
				t = 1
			}
		case hitOverride_forceair:
			f = exp[0].evalB(c)
		case hitOverride_keepstate:
			if st == -1 { // StateNo disables KeepState
				ks = exp[0].evalB(c)
			}
		case hitOverride_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	if st < 0 && !ks {
		t = 0
	}
	pn := crun.playerNo
	crun.ho[s] = HitOverride{attr: a, stateno: st, time: t, forceair: f, keepState: ks, playerNo: pn}
	return false
}

type pause StateControllerBase

const (
	pause_time byte = iota
	pause_movetime
	pause_pausebg
	pause_endcmdbuftime
	pause_redirectid
)

func (sc pause) Run(c *Char, _ []int32) bool {
	crun := c
	var t, mt int32 = 0, 0
	sys.pausebg, sys.pauseendcmdbuftime = true, 0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case pause_time:
			t = exp[0].evalI(c)
		case pause_movetime:
			mt = exp[0].evalI(c)
		case pause_pausebg:
			sys.pausebg = exp[0].evalB(c)
		case pause_endcmdbuftime:
			sys.pauseendcmdbuftime = exp[0].evalI(c)
		case pause_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.setPauseTime(t, mt)
	return false
}

type superPause StateControllerBase

const (
	superPause_time byte = iota
	superPause_movetime
	superPause_pausebg
	superPause_endcmdbuftime
	superPause_darken
	superPause_anim
	superPause_pos
	superPause_p2defmul
	superPause_poweradd
	superPause_unhittable
	superPause_sound
	superPause_redirectid
)

func (sc superPause) Run(c *Char, _ []int32) bool {
	crun := c
	var t, mt int32 = 30, 0
	uh := true
	sys.superanim, sys.superpmap.remap = crun.getAnim(100, "f", true), nil
	sys.superpos, sys.superfacing = [...]float32{crun.pos[0] * crun.localscl, crun.pos[1] * crun.localscl}, crun.facing
	sys.superpausebg, sys.superendcmdbuftime, sys.superdarken = true, 0, true
	sys.superp2defmul = crun.gi().constants["super.targetdefencemul"]
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case superPause_time:
			t = exp[0].evalI(c)
		case superPause_movetime:
			mt = exp[0].evalI(c)
		case superPause_pausebg:
			sys.superpausebg = exp[0].evalB(c)
		case superPause_endcmdbuftime:
			sys.superendcmdbuftime = exp[0].evalI(c)
		case superPause_darken:
			sys.superdarken = exp[0].evalB(c)
		case superPause_anim:
			ffx := string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			if sys.superanim = crun.getAnim(exp[1].evalI(c), ffx, true); sys.superanim != nil {
				if ffx != "" && ffx != "s" {
					sys.superpmap.remap = nil
				} else {
					sys.superpmap.remap = crun.getPalMap()
				}
			}
		case superPause_pos:
			sys.superpos[0] += crun.facing * exp[0].evalF(c) * c.localscl
			if len(exp) > 1 {
				sys.superpos[1] += exp[1].evalF(c) * c.localscl
			}
		case superPause_p2defmul:
			sys.superp2defmul = exp[0].evalF(c)
			if sys.superp2defmul == 0 {
				sys.superp2defmul = crun.gi().constants["super.targetdefencemul"]
			}
		case superPause_poweradd:
			crun.powerAdd(exp[0].evalI(c))
		case superPause_unhittable:
			uh = exp[0].evalB(c)
		case superPause_sound:
			n := int32(0)
			if len(exp) > 2 {
				n = exp[2].evalI(c)
			}
			vo := int32(100)
			ffx := string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			crun.playSound(ffx, false, 0, exp[1].evalI(c), n, -1,
				vo, 0, 1, 1, nil, false, 0, 0, 0, 0, false, false)
		case superPause_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				sys.superanim, sys.superpmap.remap = crun.getAnim(100, "f", true), nil
				sys.superpos, sys.superfacing = [...]float32{crun.pos[0] * crun.localscl, crun.pos[1] * crun.localscl}, crun.facing
			} else {
				return false
			}
		}
		return true
	})
	if sys.superanim != nil {
		sys.superanim.start_scale[0] *= crun.localscl
		sys.superanim.start_scale[1] *= crun.localscl
	}
	crun.setSuperPauseTime(t, mt, uh)
	return false
}

type trans StateControllerBase

const (
	trans_trans byte = iota
	trans_redirectid
)

func (sc trans) Run(c *Char, _ []int32) bool {
	crun := c
	crun.alpha[1] = 255
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case trans_trans:
			crun.alpha[0] = exp[0].evalI(c)
			crun.alpha[1] = exp[1].evalI(c)
			if len(exp) >= 3 {
				crun.alpha[0] = Clamp(crun.alpha[0], 0, 255)
				crun.alpha[1] = Clamp(crun.alpha[1], 0, 255)
				//if len(exp) >= 4 {
				//	crun.alpha[1] = ^crun.alpha[1]
				//} else if crun.alpha[0] == 1 && crun.alpha[1] == 255 {
				if crun.alpha[0] == 1 && crun.alpha[1] == 255 {
					crun.alpha[0] = 0
				}
			}
			crun.alphaTrg[0] = crun.alpha[0]
			crun.alphaTrg[1] = crun.alpha[1]
		case trans_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.setCSF(CSF_trans)
	return false
}

type playerPush StateControllerBase

const (
	playerPush_value byte = iota
	playerPush_redirectid
)

func (sc playerPush) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case playerPush_value:
			if exp[0].evalB(c) {
				crun.setCSF(CSF_playerpush)
			} else {
				crun.unsetCSF(CSF_playerpush)
			}
		case playerPush_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type stateTypeSet StateControllerBase

const (
	stateTypeSet_statetype byte = iota
	stateTypeSet_movetype
	stateTypeSet_physics
	stateTypeSet_redirectid
)

func (sc stateTypeSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case stateTypeSet_statetype:
			crun.ss.changeStateType(StateType(exp[0].evalI(c)))
		case stateTypeSet_movetype:
			crun.ss.changeMoveType(MoveType(exp[0].evalI(c)))
		case stateTypeSet_physics:
			crun.ss.physics = StateType(exp[0].evalI(c))
		case stateTypeSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type angleDraw StateControllerBase

const (
	angleDraw_value byte = iota
	angleDraw_scale
	angleDraw_rescaleClsn
	angleDraw_redirectid
)

func (sc angleDraw) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case angleDraw_value:
			crun.angleSet(exp[0].evalF(c))
		case angleDraw_scale:
			crun.angleScale[0] *= exp[0].evalF(c)
			crun.angleScaleTrg[0] = crun.angleScale[0]
			if len(exp) > 1 {
				crun.angleScale[1] *= exp[1].evalF(c)
				crun.angleScaleTrg[1] = crun.angleScale[1]
			}
		case angleDraw_rescaleClsn:
			if exp[0].evalB(c) {
				crun.clsnScale[0] *= crun.angleScale[0]
				crun.clsnScale[1] *= crun.angleScale[1]
				crun.angleRescaleClsn = true
			}
		case angleDraw_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.setCSF(CSF_angledraw)
	return false
}

type angleSet StateControllerBase

const (
	angleSet_value byte = iota
	angleSet_redirectid
)

func (sc angleSet) Run(c *Char, _ []int32) bool {
	crun := c
	v := float32(0) // Mugen uses 0 if no value is set at all
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case angleSet_value:
			v = exp[0].evalF(c)
		case angleSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.angleSet(v)
	return false
}

type angleAdd StateControllerBase

const (
	angleAdd_value byte = iota
	angleAdd_redirectid
)

func (sc angleAdd) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case angleAdd_value:
			crun.angleSet(crun.angle + exp[0].evalF(c))
		case angleAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type angleMul StateControllerBase

const (
	angleMul_value byte = iota
	angleMul_redirectid
)

func (sc angleMul) Run(c *Char, _ []int32) bool {
	crun := c
	v := float32(0) // Mugen uses 0 if no value is set at all
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case angleMul_value:
			v = exp[0].evalF(c)
		case angleMul_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.angleSet(crun.angle * v)
	return false
}

type envColor StateControllerBase

const (
	envColor_value byte = iota
	envColor_time
	envColor_under
)

func (sc envColor) Run(c *Char, _ []int32) bool {
	sys.envcol = [...]int32{255, 255, 255}
	sys.envcol_time = 1
	sys.envcol_under = false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case envColor_value:
			sys.envcol[0] = exp[0].evalI(c)
			sys.envcol[1] = exp[1].evalI(c)
			sys.envcol[2] = exp[2].evalI(c)
		case envColor_time:
			sys.envcol_time = exp[0].evalI(c)
		case envColor_under:
			sys.envcol_under = exp[0].evalB(c)
		}
		return true
	})
	return false
}

type displayToClipboard StateControllerBase

const (
	displayToClipboard_params byte = iota
	displayToClipboard_text
	displayToClipboard_redirectid
)

func (sc displayToClipboard) Run(c *Char, _ []int32) bool {
	crun := c
	params := []interface{}{}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case displayToClipboard_params:
			for _, e := range exp {
				if bv := e.run(c); bv.t == VT_Float {
					params = append(params, bv.ToF())
				} else {
					params = append(params, bv.ToI())
				}
			}
		case displayToClipboard_text:
			crun.clipboardText = nil
			crun.appendToClipboard(sys.workingState.playerNo,
				int(exp[0].evalI(c)), params...)
		case displayToClipboard_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type appendToClipboard displayToClipboard

func (sc appendToClipboard) Run(c *Char, _ []int32) bool {
	crun := c
	params := []interface{}{}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case displayToClipboard_params:
			for _, e := range exp {
				if bv := e.run(c); bv.t == VT_Float {
					params = append(params, bv.ToF())
				} else {
					params = append(params, bv.ToI())
				}
			}
		case displayToClipboard_text:
			crun.appendToClipboard(sys.workingState.playerNo,
				int(exp[0].evalI(c)), params...)
		case displayToClipboard_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type clearClipboard StateControllerBase

const (
	clearClipboard_ byte = iota
	clearClipboard_redirectid
)

func (sc clearClipboard) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case clearClipboard_:
			crun.clipboardText = nil
		case clearClipboard_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type makeDust StateControllerBase

const (
	makeDust_spacing byte = iota
	makeDust_pos
	makeDust_pos2
	makeDust_redirectid
)

func (sc makeDust) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case makeDust_spacing:
			s := Max(1, exp[0].evalI(c))
			if crun.time()%s != s-1 {
				return false
			}
		case makeDust_pos:
			x, y := exp[0].evalF(c), float32(0)
			if len(exp) > 1 {
				y = exp[1].evalF(c)
			}
			crun.makeDust(x-float32(crun.size.draw.offset[0]),
				y-float32(crun.size.draw.offset[1]))
		case makeDust_pos2:
			x, y := exp[0].evalF(c), float32(0)
			if len(exp) > 1 {
				y = exp[1].evalF(c)
			}
			crun.makeDust(x-float32(crun.size.draw.offset[0]),
				y-float32(crun.size.draw.offset[1]))
		case makeDust_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type attackDist StateControllerBase

const (
	attackDist_value byte = iota
	attackDist_back
	attackDist_redirectid
)

func (sc attackDist) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case attackDist_value:
			crun.attackDist[0] = exp[0].evalF(c) * lclscround
		case attackDist_back:
			crun.attackDist[1] = exp[0].evalF(c) * lclscround
		case attackDist_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type attackMulSet StateControllerBase

const (
	attackMulSet_value byte = iota
	attackMulSet_redirectid
)

func (sc attackMulSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case attackMulSet_value:
			crun.attackMul = float32(crun.gi().data.attack) * crun.ocd().attackRatio / 100 * exp[0].evalF(c)
		case attackMulSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type defenceMulSet StateControllerBase

const (
	defenceMulSet_value byte = iota
	defenceMulSet_onHit
	defenceMulSet_mulType
	defenceMulSet_redirectid
)

func (sc defenceMulSet) Run(c *Char, _ []int32) bool {
	crun := c
	var val float32 = 1
	var onHit bool = false
	var mulType int32 = 1

	// Change default behavior for Mugen chars
	if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
		onHit = true
		mulType = 0
	}

	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case defenceMulSet_value:
			val = exp[0].evalF(c)
		case defenceMulSet_onHit:
			onHit = exp[0].evalB(c)
		case defenceMulSet_mulType:
			mulType = exp[0].evalI(c)
		case defenceMulSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})

	// Apply "value" according to "mulType"
	if mulType != 0 {
		crun.customDefense = val
	} else {
		crun.customDefense = 1.0 / val
	}

	// Apply "onHit"
	crun.defenseMulDelay = onHit

	return false
}

type fallEnvShake StateControllerBase

const (
	fallEnvShake_ byte = iota
	fallEnvShake_redirectid
)

func (sc fallEnvShake) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case fallEnvShake_:
			if crun.ghv.fall.envshake_time > 0 {
				sys.envShake = EnvShake{time: crun.ghv.fall.envshake_time,
					freq:  crun.ghv.fall.envshake_freq * math.Pi / 180,
					ampl:  float32(crun.ghv.fall.envshake_ampl),
					phase: crun.ghv.fall.envshake_phase, mul: crun.ghv.fall.envshake_mul}
				sys.envShake.setDefPhase()
				crun.ghv.fall.envshake_time = 0
			}
		case fallEnvShake_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type hitFallDamage StateControllerBase

const (
	hitFallDamage_ byte = iota
	hitFallDamage_redirectid
)

func (sc hitFallDamage) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case hitFallDamage_:
			crun.hitFallDamage()
		case hitFallDamage_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type hitFallVel StateControllerBase

const (
	hitFallVel_ byte = iota
	hitFallVel_redirectid
)

func (sc hitFallVel) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case hitFallVel_:
			crun.hitFallVel()
		case hitFallVel_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type hitFallSet StateControllerBase

const (
	hitFallSet_value byte = iota
	hitFallSet_xvel
	hitFallSet_yvel
	hitFallSet_redirectid
)

func (sc hitFallSet) Run(c *Char, _ []int32) bool {
	crun := c
	f, xv, yv := int32(-1), float32(math.NaN()), float32(math.NaN())
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case hitFallSet_value:
			f = exp[0].evalI(c)
			if len(crun.ghv.hitBy) == 0 {
				return false
			}
		case hitFallSet_xvel:
			xv = exp[0].evalF(c)
		case hitFallSet_yvel:
			yv = exp[0].evalF(c)
		case hitFallSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.hitFallSet(f, xv, yv)
	return false
}

type varRangeSet StateControllerBase

const (
	varRangeSet_first byte = iota
	varRangeSet_last
	varRangeSet_value
	varRangeSet_fvalue
	varRangeSet_redirectid
)

func (sc varRangeSet) Run(c *Char, _ []int32) bool {
	crun := c
	var first, last int32 = 0, 0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case varRangeSet_first:
			first = exp[0].evalI(c)
		case varRangeSet_last:
			last = exp[0].evalI(c)
		case varRangeSet_value:
			v := exp[0].evalI(c)
			if first >= 0 && last < int32(NumVar) {
				for i := first; i <= last; i++ {
					crun.ivar[i] = v
				}
			}
		case varRangeSet_fvalue:
			fv := exp[0].evalF(c)
			if first >= 0 && last < int32(NumFvar) {
				for i := first; i <= last; i++ {
					crun.fvar[i] = fv
				}
			}
		case varRangeSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type remapPal StateControllerBase

const (
	remapPal_source byte = iota
	remapPal_dest
	remapPal_redirectid
)

func (sc remapPal) Run(c *Char, _ []int32) bool {
	crun := c
	src := [...]int32{-1, -1}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case remapPal_source:
			src[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				src[1] = exp[1].evalI(c)
			}
		case remapPal_dest:
			dst := [...]int32{exp[0].evalI(c), -1}
			if len(exp) > 1 {
				dst[1] = exp[1].evalI(c)
			}
			crun.remapPal(crun.getPalfx(), src, dst)
		case remapPal_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type stopSnd StateControllerBase

const (
	stopSnd_channel byte = iota
	stopSnd_redirectid
)

func (sc stopSnd) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case stopSnd_channel:
			if ch := Min(255, exp[0].evalI(c)); ch < 0 {
				sys.stopAllSound()
			} else if c := crun.soundChannels.Get(ch); c != nil {
				c.Stop()
			}
		case stopSnd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type sndPan StateControllerBase

const (
	sndPan_channel byte = iota
	sndPan_pan
	sndPan_abspan
	sndPan_redirectid
)

func (sc sndPan) Run(c *Char, _ []int32) bool {
	crun := c
	ch, pan, x := int32(-1), float32(0), &crun.pos[0]
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case sndPan_channel:
			ch = exp[0].evalI(c)
		case sndPan_pan:
			pan = exp[0].evalF(c)
		case sndPan_abspan:
			pan = exp[0].evalF(c)
			x = nil
		case sndPan_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				x = &crun.pos[0]
			} else {
				return false
			}
		}
		return true
	})
	if c := crun.soundChannels.Get(ch); c != nil {
		c.SetPan(pan*crun.facing, crun.localscl, x)
	}
	return false
}

type varRandom StateControllerBase

const (
	varRandom_v byte = iota
	varRandom_range
	varRandom_redirectid
)

func (sc varRandom) Run(c *Char, _ []int32) bool {
	crun := c
	var v int32
	var min, max int32 = 0, 1000
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case varRandom_v:
			v = exp[0].evalI(c)
		case varRandom_range:
			min, max = 0, exp[0].evalI(c)
			if len(exp) > 1 {
				min, max = max, exp[1].evalI(c)
			}
		case varRandom_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.varSet(v, RandI(min, max))
	return false
}

type gravity StateControllerBase

const (
	gravity_ byte = iota
	gravity_redirectid
)

func (sc gravity) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case gravity_:
			crun.gravity()
		case gravity_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type bindToParent StateControllerBase

const (
	bindToParent_time byte = iota
	bindToParent_facing
	bindToParent_pos
	bindToParent_redirectid
)

func (sc bindToParent) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	p := crun.parent()
	var x, y float32 = 0, 0
	var time int32 = 1
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case bindToParent_time:
			time = exp[0].evalI(c)
		case bindToParent_facing:
			if f := exp[0].evalI(c); f < 0 {
				crun.bindFacing = -1
			} else if f > 0 {
				crun.bindFacing = 1
			}
		case bindToParent_pos:
			x = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				y = exp[1].evalF(c) * lclscround
			}
		case bindToParent_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
				p = crun.parent()
			} else {
				return false
			}
		}
		return true
	})
	if p == nil {
		return false
	}
	crun.bindPos[0] = x
	crun.bindPos[1] = y
	crun.setBindToId(p)
	crun.setBindTime(time)
	return false
}

type bindToRoot bindToParent

func (sc bindToRoot) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	r := crun.root()
	var x, y float32 = 0, 0
	var time int32 = 1
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case bindToParent_time:
			time = exp[0].evalI(c)
		case bindToParent_facing:
			if f := exp[0].evalI(c); f < 0 {
				crun.bindFacing = -1
			} else if f > 0 {
				crun.bindFacing = 1
			}
		case bindToParent_pos:
			x = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				y = exp[1].evalF(c) * lclscround
			}
		case bindToParent_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
				r = crun.root()
			} else {
				return false
			}
		}
		return true
	})
	if r == nil {
		return false
	}
	crun.bindPos[0] = x
	crun.bindPos[1] = y
	crun.setBindToId(r)
	crun.setBindTime(time)
	return false
}

type removeExplod StateControllerBase

const (
	removeExplod_id byte = iota
	removeExplod_redirectid
)

func (sc removeExplod) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case removeExplod_id:
			crun.removeExplod(exp[0].evalI(c))
		case removeExplod_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type explodBindTime StateControllerBase

const (
	explodBindTime_id byte = iota
	explodBindTime_time
	explodBindTime_redirectid
)

func (sc explodBindTime) Run(c *Char, _ []int32) bool {
	crun := c
	var eid, time int32 = -1, 0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case explodBindTime_id:
			eid = exp[0].evalI(c)
		case explodBindTime_time:
			time = exp[0].evalI(c)
		case explodBindTime_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.explodBindTime(eid, time)
	return false
}

type moveHitReset StateControllerBase

const (
	moveHitReset_ byte = iota
	moveHitReset_redirectid
)

func (sc moveHitReset) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case moveHitReset_:
			crun.clearMoveHit()
		case moveHitReset_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type hitAdd StateControllerBase

const (
	hitAdd_value byte = iota
	hitAdd_redirectid
)

func (sc hitAdd) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case hitAdd_value:
			crun.hitAdd(exp[0].evalI(c))
		case hitAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type offset StateControllerBase

const (
	offset_x byte = iota
	offset_y
	offset_redirectid
)

func (sc offset) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case offset_x:
			crun.offset[0] = exp[0].evalF(c) * c.localscl
			crun.offsetTrg[0] = crun.offset[0]
		case offset_y:
			crun.offset[1] = exp[0].evalF(c) * c.localscl
			crun.offsetTrg[1] = crun.offset[1]
		case offset_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type victoryQuote StateControllerBase

const (
	victoryQuote_value byte = iota
	victoryQuote_redirectid
)

func (sc victoryQuote) Run(c *Char, _ []int32) bool {
	crun := c
	var v int32 = -1
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case victoryQuote_value:
			v = exp[0].evalI(c)
		case victoryQuote_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.winquote = v
	return false
}

type zoom StateControllerBase

const (
	zoom_pos byte = iota
	zoom_scale
	zoom_lag
	zoom_redirectid
	zoom_camerabound
	zoom_time
	zoom_stagebound
)

func (sc zoom) Run(c *Char, _ []int32) bool {
	crun := c
	zoompos := [2]float32{0, 0}
	t := int32(1)
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case zoom_pos:
			zoompos[0] = exp[0].evalF(c) * crun.localscl
			if len(exp) > 1 {
				zoompos[1] = exp[1].evalF(c) * crun.localscl
			}
		case zoom_scale:
			sys.zoomScale = exp[0].evalF(c)
		case zoom_camerabound:
			sys.zoomCameraBound = exp[0].evalB(c)
		case zoom_stagebound:
			sys.zoomStageBound = exp[0].evalB(c)
		case zoom_lag:
			sys.zoomlag = exp[0].evalF(c)
		case zoom_time:
			t = exp[0].evalI(c)
		case zoom_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	sys.zoomPos[0] = sys.zoomScale * zoompos[0]
	sys.zoomPos[1] = zoompos[1]
	sys.enableZoomtime = t
	return false
}

type forceFeedback StateControllerBase

const (
	forceFeedback_waveform byte = iota
	forceFeedback_time
	forceFeedback_freq
	forceFeedback_ampl
	forceFeedback_self
	forceFeedback_redirectid
)

func (sc forceFeedback) Run(c *Char, _ []int32) bool {
	/*crun := c
	waveform := int32(0)
	time := int32(60)
	freq := [4]float32{128, 0, 0, 0}
	ampl := [4]float32{128, 0, 0, 0}
	self := true
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case forceFeedback_waveform:
			waveform = exp[0].evalI(c)
		case forceFeedback_time:
			time = exp[0].evalI(c)
		case forceFeedback_freq:
			freq[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				freq[1] = exp[1].evalF(c)
			}
			if len(exp) > 2 {
				freq[2] = exp[2].evalF(c)
			}
			if len(exp) > 3 {
				freq[3] = exp[3].evalF(c)
			}
		case forceFeedback_ampl:
			ampl[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				ampl[1] = exp[1].evalF(c)
			}
			if len(exp) > 2 {
				ampl[2] = exp[2].evalF(c)
			}
			if len(exp) > 3 {
				ampl[3] = exp[3].evalF(c)
			}
		case forceFeedback_self:
			self = exp[0].evalB(c)
		case forceFeedback_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})*/
	//TODO: not implemented
	return false
}

type assertCommand StateControllerBase

const (
	assertCommand_name byte = iota
	assertCommand_buffertime
	assertCommand_redirectid
)

func (sc assertCommand) Run(c *Char, _ []int32) bool {
	crun := c
	n := ""
	bt := int32(1)
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case assertCommand_name:
			n = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		case assertCommand_buffertime:
			bt = exp[0].evalI(c)
		case assertCommand_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.assertCommand(n, bt)
	return false
}

type assertInput StateControllerBase

const (
	assertInput_flag byte = iota
	assertInput_redirectid
)

func (sc assertInput) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case assertInput_flag:
			crun.inputFlag |= InputBits(exp[0].evalI(c))
		case assertInput_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type dialogue StateControllerBase

const (
	dialogue_hidebars byte = iota
	dialogue_force
	dialogue_text
	dialogue_redirectid
)

func (sc dialogue) Run(c *Char, _ []int32) bool {
	crun := c
	reset := true
	force := false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case dialogue_hidebars:
			sys.dialogueBarsFlg = sys.lifebar.hidebars && exp[0].evalB(c)
		case dialogue_force:
			force = exp[0].evalB(c)
		case dialogue_text:
			sys.chars[crun.playerNo][0].appendDialogue(string(*(*[]byte)(unsafe.Pointer(&exp[0]))), reset)
			reset = false
		case dialogue_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	if force {
		sys.dialogueFlg = true
		sys.dialogueForce = crun.playerNo + 1
	}
	return false
}

type dizzyPointsAdd StateControllerBase

const (
	dizzyPointsAdd_absolute byte = iota
	dizzyPointsAdd_value
	dizzyPointsAdd_redirectid
)

func (sc dizzyPointsAdd) Run(c *Char, _ []int32) bool {
	a := false
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case dizzyPointsAdd_absolute:
			a = exp[0].evalB(c)
		case dizzyPointsAdd_value:
			crun.dizzyPointsAdd(float64(exp[0].evalI(c)), a)
		case dizzyPointsAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type dizzyPointsSet StateControllerBase

const (
	dizzyPointsSet_value byte = iota
	dizzyPointsSet_redirectid
)

func (sc dizzyPointsSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case dizzyPointsSet_value:
			crun.dizzyPointsSet(exp[0].evalI(c))
		case dizzyPointsSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type dizzySet StateControllerBase

const (
	dizzySet_value byte = iota
	dizzySet_redirectid
)

func (sc dizzySet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case dizzySet_value:
			crun.setDizzy(exp[0].evalB(c))
		case dizzySet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type guardBreakSet StateControllerBase

const (
	guardBreakSet_value byte = iota
	guardBreakSet_redirectid
)

func (sc guardBreakSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case guardBreakSet_value:
			crun.setGuardBreak(exp[0].evalB(c))
		case guardBreakSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type guardPointsAdd StateControllerBase

const (
	guardPointsAdd_absolute byte = iota
	guardPointsAdd_value
	guardPointsAdd_redirectid
)

func (sc guardPointsAdd) Run(c *Char, _ []int32) bool {
	a := false
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case guardPointsAdd_absolute:
			a = exp[0].evalB(c)
		case guardPointsAdd_value:
			crun.guardPointsAdd(float64(exp[0].evalI(c)), a)
		case guardPointsAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type guardPointsSet StateControllerBase

const (
	guardPointsSet_value byte = iota
	guardPointsSet_redirectid
)

func (sc guardPointsSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case guardPointsSet_value:
			crun.guardPointsSet(exp[0].evalI(c))
		case guardPointsSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type hitScaleSet StateControllerBase

const (
	hitScaleSet_id byte = iota
	hitScaleSet_affects_damage
	hitScaleSet_affects_hitTime
	hitScaleSet_affects_pauseTime
	hitScaleSet_mul
	hitScaleSet_add
	hitScaleSet_addType
	hitScaleSet_min
	hitScaleSet_max
	hitScaleSet_time
	hitScaleSet_reset
	hitScaleSet_force
	hitScaleSet_redirectid
)

// Takes the values given by Compiler.hitScaleSet and executes it.
func (sc hitScaleSet) Run(c *Char, _ []int32) bool {
	var crun = c
	// Default values
	var affects = []bool{false, false, false}
	// Target of the hitScale, -1 is default.
	var target int32 = -1
	var targetArray [3]*HitScale
	// Do we reset everithng back to default?
	var resetAll = false
	var reset = false
	// If false we wait to hit to apply hitScale.
	// If true we apply on call.
	var force = false
	// Holder variables
	var tempHitScale = newHitScale()

	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		// What is hitScale ging to affect.
		case hitScaleSet_affects_damage:
			affects[0] = true
		case hitScaleSet_affects_hitTime:
			affects[1] = true
		case hitScaleSet_affects_pauseTime:
			affects[2] = true
		// ID of the char to apply to.
		case hitScaleSet_id:
			target = exp[0].evalI(c)
		case hitScaleSet_mul:
			tempHitScale.mul = exp[0].evalF(c)
		case hitScaleSet_add:
			tempHitScale.add = exp[0].evalI(c)
		case hitScaleSet_addType:
			tempHitScale.addType = exp[0].evalI(c)
		case hitScaleSet_min:
			tempHitScale.min = exp[0].evalF(c)
		case hitScaleSet_max:
			tempHitScale.max = exp[0].evalF(c)
		case hitScaleSet_time:
			tempHitScale.time = exp[0].evalI(c)
		case hitScaleSet_reset:
			if exp[0].evalI(c) == 1 {
				reset = true
			} else if exp[0].evalI(c) == 2 {
				resetAll = true
			}
		case hitScaleSet_force:
			force = exp[0].evalB(c)
		// Genric redirectId.
		case hitScaleSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})

	// ----------------------------------------------------------------------

	if resetAll {
		for _, hs := range crun.defaultHitScale {
			hs.reset()
		}
		crun.nextHitScale = make(map[int32][3]*HitScale)
		crun.activeHitScale = make(map[int32][3]*HitScale)
	}

	targetArray = getHitScaleTarget(crun, target, force, reset)

	// Apply the new values and activate it.
	for i, hs := range targetArray {
		if affects[i] {
			if reset {
				if ahs, ok := crun.activeHitScale[target]; ok {
					ahs[int32(i)].reset()
				}
			}
			hs.copy(tempHitScale)
			hs.active = true
		}
	}

	return false
}

func getHitScaleTarget(char *Char, target int32, force bool, reset bool) [3]*HitScale {
	// Get our targets.
	if target <= -1 {
		return char.defaultHitScale
	} else { //Check if target exists.
		if force {
			if _, ok := char.activeHitScale[target]; !ok || reset {
				char.activeHitScale[target] = newHitScaleArray()
			}
			return char.activeHitScale[target]
		} else {
			if _, ok := char.nextHitScale[target]; !ok || reset {
				char.nextHitScale[target] = newHitScaleArray()
			}
			return char.nextHitScale[target]
		}
	}
}

type lifebarAction StateControllerBase

const (
	lifebarAction_top byte = iota
	lifebarAction_time
	lifebarAction_timemul
	lifebarAction_anim
	lifebarAction_spr
	lifebarAction_snd
	lifebarAction_text
	lifebarAction_redirectid
)

func (sc lifebarAction) Run(c *Char, _ []int32) bool {
	crun := c
	var top bool
	var text string
	var timemul float32 = 1
	var time, anim int32 = -1, -1
	spr := [2]int32{-1, 0}
	snd := [2]int32{-1, 0}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case lifebarAction_top:
			top = exp[0].evalB(c)
		case lifebarAction_timemul:
			timemul = exp[0].evalF(c)
		case lifebarAction_time:
			time = exp[0].evalI(c)
		case lifebarAction_anim:
			anim = exp[0].evalI(c)
		case lifebarAction_spr:
			spr[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				spr[1] = exp[1].evalI(c)
			}
		case lifebarAction_snd:
			snd[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				snd[1] = exp[1].evalI(c)
			}
		case lifebarAction_text:
			text = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		case lifebarAction_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.appendLifebarAction(text, snd, spr, anim, time, timemul, top)
	return false
}

type loadFile StateControllerBase

const (
	loadFile_path byte = iota
	loadFile_saveData
	loadFile_redirectid
)

func (sc loadFile) Run(c *Char, _ []int32) bool {
	crun := c
	var path string
	var data SaveData
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case loadFile_path:
			path = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		case loadFile_saveData:
			data = SaveData(exp[0].evalI(c))
		case loadFile_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	if path != "" {
		decodeFile, err := os.Open(filepath.Dir(c.gi().def) + "/" + path)
		if err != nil {
			defer decodeFile.Close()
			return false
		}
		defer decodeFile.Close()
		decoder := gob.NewDecoder(decodeFile)
		switch data {
		case SaveData_map:
			if err := decoder.Decode(&crun.mapArray); err != nil {
				panic(err)
			}
		case SaveData_var:
			if err := decoder.Decode(&crun.ivar); err != nil {
				panic(err)
			}
		case SaveData_fvar:
			if err := decoder.Decode(&crun.fvar); err != nil {
				panic(err)
			}
		}
	}
	return false
}

type mapSet StateControllerBase

const (
	mapSet_mapArray byte = iota
	mapSet_value
	mapSet_redirectid
	mapSet_type
)

func (sc mapSet) Run(c *Char, _ []int32) bool {
	crun := c
	var s string
	var value float32
	var scType int32
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case mapSet_mapArray:
			s = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		case mapSet_value:
			value = exp[0].evalF(c)
		case mapSet_type:
			scType = exp[0].evalI(c)
		case mapSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.mapSet(s, value, scType)
	return false
}

type matchRestart StateControllerBase

const (
	matchRestart_reload byte = iota
	matchRestart_stagedef
	matchRestart_p1def
	matchRestart_p2def
	matchRestart_p3def
	matchRestart_p4def
	matchRestart_p5def
	matchRestart_p6def
	matchRestart_p7def
	matchRestart_p8def
)

func (sc matchRestart) Run(c *Char, _ []int32) bool {
	var s string
	reloadFlag := false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case matchRestart_reload:
			for i, p := range exp {
				sys.reloadCharSlot[i] = p.evalB(c)
				if sys.reloadCharSlot[i] {
					reloadFlag = true
				}
			}
		case matchRestart_stagedef:
			s = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			sys.sel.sdefOverwrite = SearchFile(s, []string{c.gi().def})
			//sys.reloadStageFlg = true
			reloadFlag = true
		case matchRestart_p1def:
			s = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			sys.sel.cdefOverwrite[0] = SearchFile(s, []string{c.gi().def})
		case matchRestart_p2def:
			s = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			sys.sel.cdefOverwrite[1] = SearchFile(s, []string{c.gi().def})
		case matchRestart_p3def:
			s = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			sys.sel.cdefOverwrite[2] = SearchFile(s, []string{c.gi().def})
		case matchRestart_p4def:
			s = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			sys.sel.cdefOverwrite[3] = SearchFile(s, []string{c.gi().def})
		case matchRestart_p5def:
			s = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			sys.sel.cdefOverwrite[4] = SearchFile(s, []string{c.gi().def})
		case matchRestart_p6def:
			s = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			sys.sel.cdefOverwrite[5] = SearchFile(s, []string{c.gi().def})
		case matchRestart_p7def:
			s = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			sys.sel.cdefOverwrite[6] = SearchFile(s, []string{c.gi().def})
		case matchRestart_p8def:
			s = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			sys.sel.cdefOverwrite[7] = SearchFile(s, []string{c.gi().def})
		}
		return true
	})
	if sys.netInput == nil && sys.fileInput == nil {
		if reloadFlag {
			sys.reloadFlg = true
		} else {
			sys.roundResetFlg = true
		}
	}
	return false
}

type printToConsole StateControllerBase

const (
	printToConsole_params byte = iota
	printToConsole_text
)

func (sc printToConsole) Run(c *Char, _ []int32) bool {
	params := []interface{}{}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case printToConsole_params:
			for _, e := range exp {
				if bv := e.run(c); bv.t == VT_Float {
					params = append(params, bv.ToF())
				} else {
					params = append(params, bv.ToI())
				}
			}
		case printToConsole_text:
			sys.printToConsole(sys.workingState.playerNo,
				int(exp[0].evalI(c)), params...)
		}
		return true
	})
	return false
}

type redLifeAdd StateControllerBase

const (
	redLifeAdd_absolute byte = iota
	redLifeAdd_value
	redLifeAdd_redirectid
)

func (sc redLifeAdd) Run(c *Char, _ []int32) bool {
	a := false
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case redLifeAdd_absolute:
			a = exp[0].evalB(c)
		case redLifeAdd_value:
			crun.redLifeAdd(float64(exp[0].evalI(c)), a)
		case redLifeAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type redLifeSet StateControllerBase

const (
	redLifeSet_value byte = iota
	redLifeSet_redirectid
)

func (sc redLifeSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case redLifeSet_value:
			crun.redLifeSet(exp[0].evalI(c))
		case redLifeSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type remapSprite StateControllerBase

const (
	remapSprite_reset byte = iota
	remapSprite_preset
	remapSprite_source
	remapSprite_dest
	remapSprite_redirectid
)

func (sc remapSprite) Run(c *Char, _ []int32) bool {
	crun := c
	src := [...]int16{-1, -1}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case remapSprite_reset:
			if exp[0].evalB(c) {
				crun.remapSpr = make(RemapPreset)
			}
		case remapSprite_preset:
			crun.remapSpritePreset(string(*(*[]byte)(unsafe.Pointer(&exp[0]))))
		case remapSprite_source:
			src[0] = int16(exp[0].evalI(c))
			if len(exp) > 1 {
				src[1] = int16(exp[1].evalI(c))
			}
		case remapSprite_dest:
			dst := [...]int16{int16(exp[0].evalI(c)), -1}
			if len(exp) > 1 {
				dst[1] = int16(exp[1].evalI(c))
			}
			crun.remapSprite(src, dst)
		case remapSprite_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.anim.remap = crun.remapSpr
	return false
}

type roundTimeAdd StateControllerBase

const (
	roundTimeAdd_value byte = iota
	roundTimeAdd_redirectid
)

func (sc roundTimeAdd) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case roundTimeAdd_value:
			if sys.roundTime != -1 {
				sys.time = Clamp(sys.time+exp[0].evalI(c), 0, sys.roundTime)
			}
		}
		return true
	})
	return false
}

type roundTimeSet StateControllerBase

const (
	roundTimeSet_value byte = iota
	roundTimeSet_redirectid
)

func (sc roundTimeSet) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case roundTimeSet_value:
			if sys.roundTime != -1 {
				sys.time = Clamp(exp[0].evalI(c), 0, sys.roundTime)
			}
		}
		return true
	})
	return false
}

type saveFile StateControllerBase

const (
	saveFile_path byte = iota
	saveFile_saveData
	saveFile_redirectid
)

func (sc saveFile) Run(c *Char, _ []int32) bool {
	crun := c
	var path string
	var data SaveData
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case saveFile_path:
			path = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		case saveFile_saveData:
			data = SaveData(exp[0].evalI(c))
		case saveFile_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	if path != "" {
		encodeFile, err := os.Create(filepath.Dir(c.gi().def) + "/" + path)
		if err != nil {
			panic(err)
		}
		defer encodeFile.Close()
		encoder := gob.NewEncoder(encodeFile)
		switch data {
		case SaveData_map:
			if err := encoder.Encode(crun.mapArray); err != nil {
				panic(err)
			}
		case SaveData_var:
			if err := encoder.Encode(crun.ivar); err != nil {
				panic(err)
			}
		case SaveData_fvar:
			if err := encoder.Encode(crun.fvar); err != nil {
				panic(err)
			}
		}
	}
	return false
}

type scoreAdd StateControllerBase

const (
	scoreAdd_value byte = iota
	scoreAdd_redirectid
)

func (sc scoreAdd) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case scoreAdd_value:
			crun.scoreAdd(exp[0].evalF(c))
		case scoreAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type modifyBGCtrl StateControllerBase

const (
	modifyBGCtrl_id byte = iota
	modifyBGCtrl_time
	modifyBGCtrl_value
	modifyBGCtrl_x
	modifyBGCtrl_y
	modifyBGCtrl_source
	modifyBGCtrl_dest
	modifyBGCtrl_add
	modifyBGCtrl_mul
	modifyBGCtrl_sinadd
	modifyBGCtrl_sinmul
	modifyBGCtrl_sincolor
	modifyBGCtrl_sinhue
	modifyBGCtrl_invertall
	modifyBGCtrl_invertblend
	modifyBGCtrl_color
	modifyBGCtrl_hue
	modifyBGCtrl_redirectid
)

func (sc modifyBGCtrl) Run(c *Char, _ []int32) bool {
	//crun := c
	var cid int32
	t, v := [3]int32{IErr, IErr, IErr}, [3]int32{IErr, IErr, IErr}
	x, y := float32(math.NaN()), float32(math.NaN())
	src, dst := [2]int32{IErr, IErr}, [2]int32{IErr, IErr}
	add, mul, sinadd, sinmul, sincolor, sinhue := [3]int32{IErr, IErr, IErr}, [3]int32{IErr, IErr, IErr}, [4]int32{IErr, IErr, IErr, IErr}, [4]int32{IErr, IErr, IErr, IErr}, [2]int32{IErr, IErr}, [2]int32{IErr, IErr}
	invall, invblend, color, hue := IErr, IErr, float32(math.NaN()), float32(math.NaN())
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case modifyBGCtrl_id:
			cid = exp[0].evalI(c)
		case modifyBGCtrl_time:
			t[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				t[1] = exp[1].evalI(c)
				if len(exp) > 2 {
					t[2] = exp[2].evalI(c)
				}
			}
		case modifyBGCtrl_value:
			v[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				v[1] = exp[1].evalI(c)
				if len(exp) > 2 {
					v[2] = exp[2].evalI(c)
				}
			}
		case modifyBGCtrl_x:
			x = exp[0].evalF(c)
		case modifyBGCtrl_y:
			y = exp[0].evalF(c)
		case modifyBGCtrl_source:
			src[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				src[1] = exp[1].evalI(c)
			}
		case modifyBGCtrl_dest:
			dst[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				dst[1] = exp[1].evalI(c)
			}
		case modifyBGCtrl_add:
			add[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				add[1] = exp[1].evalI(c)
				if len(exp) > 2 {
					add[2] = exp[2].evalI(c)
				}
			}
		case modifyBGCtrl_mul:
			mul[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				mul[1] = exp[1].evalI(c)
				if len(exp) > 2 {
					mul[2] = exp[2].evalI(c)
				}
			}
		case modifyBGCtrl_sinadd:
			sinadd[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				sinadd[1] = exp[1].evalI(c)
				if len(exp) > 2 {
					sinadd[2] = exp[2].evalI(c)
					if len(exp) > 3 {
						sinadd[3] = exp[3].evalI(c)
					}
				}
			}
		case modifyBGCtrl_sinmul:
			sinmul[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				sinmul[1] = exp[1].evalI(c)
				if len(exp) > 2 {
					sinmul[2] = exp[2].evalI(c)
					if len(exp) > 3 {
						sinmul[3] = exp[3].evalI(c)
					}
				}
			}
		case modifyBGCtrl_sincolor:
			sincolor[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				sincolor[1] = exp[1].evalI(c)
			}
		case modifyBGCtrl_sinhue:
			sinhue[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				sinhue[1] = exp[1].evalI(c)
			}
		case modifyBGCtrl_invertall:
			invall = exp[0].evalI(c)
		case modifyBGCtrl_invertblend:
			invblend = exp[0].evalI(c)
		case modifyBGCtrl_color:
			color = exp[0].evalF(c)
		case modifyBGCtrl_hue:
			hue = exp[0].evalF(c)
		case modifyBGCtrl_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				//crun = rid
			} else {
				return false
			}
		}
		return true
	})
	sys.stage.modifyBGCtrl(cid, t, v, x, y, src, dst, add, mul, sinadd, sinmul, sincolor, sinhue, invall, invblend, color, hue)
	return false
}

type modifyBgm StateControllerBase

const (
	modifyBgm_volume = iota
	modifyBgm_loopstart
	modifyBgm_loopend
	modifyBgm_position
	modifyBgm_freqmul
	modifyBgm_redirectid
)

func (sc modifyBgm) Run(c *Char, _ []int32) bool {
	var volumeSet, loopStartSet, loopEndSet, posSet, freqSet = false, false, false, false, false
	var volume, loopstart, loopend, position int = 100, 0, 0, 0
	var freqmul float32 = 1.0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case modifyBgm_volume:
			volume = int(exp[0].evalI(c))
			volumeSet = true
		case modifyBgm_loopstart:
			loopstart = int(exp[0].evalI64(c))
			loopStartSet = true
		case modifyBgm_loopend:
			loopend = int(exp[0].evalI64(c))
			loopEndSet = true
		case modifyBgm_position:
			position = int(exp[0].evalI64(c))
			posSet = true
		case modifyBgm_freqmul:
			freqmul = float32(exp[0].evalF(c))
			freqSet = true
		case modifyBgm_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {

			} else {
				return false
			}
		}
		return true
	})
	if sys.bgm.ctrl != nil {
		// Set values that are different only
		if volumeSet {
			volumeScaled := int(float64(volume) / 100.0 * float64(sys.maxBgmVolume))
			sys.bgm.bgmVolume = int(Min(int32(volumeScaled), int32(sys.maxBgmVolume)))
			sys.bgm.UpdateVolume()
		}
		if posSet {
			sys.bgm.Seek(position)
		}
		if sl, ok := sys.bgm.volctrl.Streamer.(*StreamLooper); ok {
			if (loopStartSet && sl.loopstart != loopstart) || (loopEndSet && sl.loopend != loopend) {
				sys.bgm.SetLoopPoints(loopstart, loopend)
			}
		}
		if freqSet && sys.bgm.freqmul != freqmul {
			sys.bgm.SetFreqMul(freqmul)
		}
	}
	return false
}

type modifySnd StateControllerBase

const (
	modifySnd_channel = iota
	modifySnd_pan
	modifySnd_abspan
	modifySnd_volume
	modifySnd_volumescale
	modifySnd_freqmul
	modifySnd_redirectid
	modifySnd_priority
	modifySnd_loopstart
	modifySnd_loopend
	modifySnd_position
	modifySnd_loop
	modifySnd_loopcount
	modifySnd_stopongethit
	modifySnd_stoponchangestate
)

func (sc modifySnd) Run(c *Char, _ []int32) bool {
	if sys.noSoundFlg {
		return false
	}
	crun := c
	snd := crun.soundChannels.Get(-1)
	var ch, pri int32 = -1, 0
	var vo, fr float32 = 100, 1.0
	stopgh, stopcs := false, false
	freqMulSet, volumeSet, prioritySet, panSet, loopStartSet, loopEndSet, posSet, lcSet, loopSet := false, false, false, false, false, false, false, false, false
	stopghSet, stopcsSet := false, false
	var loopstart, loopend, position, lc int = 0, 0, 0, 0
	var p float32 = 0
	x := &c.pos[0]
	ls := crun.localscl
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case modifySnd_channel:
			ch = exp[0].evalI(c)
		case modifySnd_pan:
			p = exp[0].evalF(c)
			panSet = true
		case modifySnd_abspan:
			x = nil
			ls = 1
			p = exp[0].evalF(c)
			panSet = true
		case modifySnd_volume:
			vo = (vo + float32(exp[0].evalI(c))*(25.0/64.0)) * (64.0 / 25.0)
			volumeSet = true
		case modifySnd_volumescale:
			vo = float32(crun.gi().data.volume * exp[0].evalI(c) / 100)
			volumeSet = true
		case modifySnd_freqmul:
			fr = ClampF(exp[0].evalF(c), 0.01, 5)
			freqMulSet = true
		case modifySnd_priority:
			pri = exp[0].evalI(c)
			prioritySet = true
		case modifySnd_loopstart:
			loopstart = int(exp[0].evalI64(c))
			loopStartSet = true
		case modifySnd_loopend:
			loopend = int(exp[0].evalI64(c))
			loopEndSet = true
		case modifySnd_position:
			position = int(exp[0].evalI64(c))
			posSet = true
		case modifySnd_loop:
			if lc == 0 {
				if bool(exp[0].evalB(c)) {
					lc = -1
				} else {
					lc = 0
				}
				loopSet = true
			}
		case modifySnd_loopcount:
			lc = int(exp[0].evalI(c))
			lcSet = true
		case modifySnd_stopongethit:
			stopgh = exp[0].evalB(c)
		case modifySnd_stoponchangestate:
			stopcs = exp[0].evalB(c)
		case modifySnd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				x = &crun.pos[0]
				ls = crun.localscl
				snd = crun.soundChannels.Get(ch)
			} else {
				return false
			}
		}
		return true
	})
	// Grab the correct sound channel now
	channelCount := 1
	if ch < 0 {
		channelCount = len(crun.soundChannels.channels)
	}
	for i := channelCount - 1; i >= 0; i-- {
		if ch < 0 {
			snd = &crun.soundChannels.channels[i]
		} else {
			snd = crun.soundChannels.Get(ch)
		}

		if snd != nil && snd.sfx != nil {
			// If we didn't set the values, default them to current values.
			if !freqMulSet {
				fr = snd.sfx.freqmul
			}
			if !volumeSet {
				vo = snd.sfx.volume
			}
			if !prioritySet {
				pri = snd.sfx.priority
			}
			if !panSet {
				p = snd.sfx.p
				ls = snd.sfx.ls
				x = snd.sfx.x
			}

			// Now set the values if they're different
			if snd.sfx.freqmul != fr {
				snd.SetFreqMul(fr)
			}
			if pri != snd.sfx.priority {
				snd.SetPriority(pri)
			}
			if posSet {
				snd.streamer.Seek(position)
			}
			if lcSet || loopSet {
				if sl, ok := snd.sfx.streamer.(*StreamLooper); ok {
					sl.loopcount = lc
				}
			}
			if sl, ok := snd.sfx.streamer.(*StreamLooper); ok {
				if (loopStartSet && sl.loopstart != loopstart) || (loopEndSet && sl.loopend != loopend) {
					snd.SetLoopPoints(loopstart, loopend)
				}
			}
			if p != snd.sfx.p || ls != snd.sfx.ls || x != snd.sfx.x {
				snd.SetPan(p*crun.facing, ls, x)
			}
			if vo != snd.sfx.volume {
				snd.SetVolume(vo)
			}
			// These flags can be updated regardless since there are no calculations involved
			if stopghSet {
				snd.stopOnGetHit = stopgh
			}
			if stopcsSet {
				snd.stopOnChangeState = stopcs
			}
		}
	}
	return false
}

type playBgm StateControllerBase

const (
	playBgm_bgm = iota
	playBgm_volume
	playBgm_loop
	playBgm_loopstart
	playBgm_loopend
	playBgm_startposition
	playBgm_freqmul
	playBgm_redirectid
)

func (sc playBgm) Run(c *Char, _ []int32) bool {
	crun := c
	var b bool
	var bgm string
	var loop, volume, loopstart, loopend, startposition int = 1, 100, 0, 0, 0
	var freqmul float32 = 1.0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case playBgm_bgm:
			if bgm = string(*(*[]byte)(unsafe.Pointer(&exp[0]))); bgm != "" {
				bgm = SearchFile(bgm, []string{crun.gi().def, "", "sound/"})
			}
			b = true
		case playBgm_volume:
			volume = int(exp[0].evalI(c))
			if !b {
				sys.bgm.bgmVolume = int(Min(int32(volume), int32(sys.maxBgmVolume)))
				sys.bgm.UpdateVolume()
			}
		case playBgm_loop:
			loop = int(exp[0].evalI(c))
		case playBgm_loopstart:
			loopstart = int(exp[0].evalI(c))
		case playBgm_loopend:
			loopend = int(exp[0].evalI(c))
		case playBgm_startposition:
			startposition = int(exp[0].evalI(c))
		case playBgm_freqmul:
			freqmul = exp[0].evalF(c)
		case playBgm_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	if b {
		sys.bgm.Open(bgm, loop, volume, loopstart, loopend, startposition, freqmul)
		sys.playBgmFlg = true
	}
	return false
}

type targetDizzyPointsAdd StateControllerBase

const (
	targetDizzyPointsAdd_id byte = iota
	targetDizzyPointsAdd_absolute
	targetDizzyPointsAdd_value
	targetDizzyPointsAdd_redirectid
)

func (sc targetDizzyPointsAdd) Run(c *Char, _ []int32) bool {
	crun := c
	tar, a := crun.getTarget(-1), false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetDizzyPointsAdd_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case targetDizzyPointsAdd_absolute:
			a = exp[0].evalB(c)
		case targetDizzyPointsAdd_value:
			if len(tar) == 0 {
				return false
			}
			crun.targetDizzyPointsAdd(tar, exp[0].evalI(c), a)
		case targetDizzyPointsAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type targetGuardPointsAdd StateControllerBase

const (
	targetGuardPointsAdd_id byte = iota
	targetGuardPointsAdd_absolute
	targetGuardPointsAdd_value
	targetGuardPointsAdd_redirectid
)

func (sc targetGuardPointsAdd) Run(c *Char, _ []int32) bool {
	crun := c
	tar, a := crun.getTarget(-1), false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetGuardPointsAdd_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case targetGuardPointsAdd_absolute:
			a = exp[0].evalB(c)
		case targetGuardPointsAdd_value:
			if len(tar) == 0 {
				return false
			}
			crun.targetGuardPointsAdd(tar, exp[0].evalI(c), a)
		case targetGuardPointsAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type targetRedLifeAdd StateControllerBase

const (
	targetRedLifeAdd_id byte = iota
	targetRedLifeAdd_absolute
	targetRedLifeAdd_value
	targetRedLifeAdd_redirectid
)

func (sc targetRedLifeAdd) Run(c *Char, _ []int32) bool {
	crun := c
	tar, a := crun.getTarget(-1), false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetRedLifeAdd_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case targetRedLifeAdd_absolute:
			a = exp[0].evalB(c)
		case targetRedLifeAdd_value:
			if len(tar) == 0 {
				return false
			}
			v := exp[0].evalI(c)
			// Mugen forces absolute parameter when healing characters
			if v > 0 && c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
				a = true
			}
			crun.targetRedLifeAdd(tar, exp[0].evalI(c), a)
		case targetRedLifeAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type targetScoreAdd StateControllerBase

const (
	targetScoreAdd_id byte = iota
	targetScoreAdd_value
	targetScoreAdd_redirectid
)

func (sc targetScoreAdd) Run(c *Char, _ []int32) bool {
	crun := c
	tar := crun.getTarget(-1)
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetScoreAdd_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case targetScoreAdd_value:
			if len(tar) == 0 {
				return false
			}
			crun.targetScoreAdd(tar, exp[0].evalF(c))
		case targetScoreAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type text StateControllerBase

const (
	text_removetime byte = iota
	text_layerno
	text_params
	text_font
	text_localcoord
	text_bank
	text_align
	text_text
	text_pos
	text_scale
	text_color
	text_redirectid
)

func (sc text) Run(c *Char, _ []int32) bool {
	crun := c
	params := []interface{}{}
	ts := NewTextSprite()
	ts.SetLocalcoord(float32(sys.scrrect[2]), float32(sys.scrrect[3]))
	var xscl, yscl float32 = 1, 1
	var fnt int = -1
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case text_removetime:
			ts.removetime = exp[0].evalI(c)
		case text_layerno:
			ts.layerno = int16(exp[0].evalI(c))
		case text_params:
			for _, e := range exp {
				if bv := e.run(c); bv.t == VT_Float {
					params = append(params, bv.ToF())
				} else {
					params = append(params, bv.ToI())
				}
			}
		case text_text:
			sn := int(exp[0].evalI(c))
			spl := sys.stringPool[sys.workingState.playerNo].List
			if sn >= 0 && sn < len(spl) {
				ts.text = OldSprintf(spl[sn], params...)
			}
		case text_font:
			fnt = int(exp[1].evalI(c))
			fflg := exp[0].evalB(c)
			fntList := crun.gi().fnt
			if fflg {
				fntList = sys.lifebar.fnt
			}
			if fnt >= 0 && fnt < len(fntList) && fntList[fnt] != nil {
				ts.fnt = fntList[fnt]
				if fflg {
					ts.SetLocalcoord(float32(sys.lifebarLocalcoord[0]), float32(sys.lifebarLocalcoord[1]))
				} else {
					//ts.SetLocalcoord(c.stOgi().localcoord[0], c.stOgi().localcoord[1])
				}
			} else {
				fnt = -1
			}
		case text_localcoord:
			ts.SetLocalcoord(exp[0].evalF(c), exp[1].evalF(c))
		case text_bank:
			ts.bank = exp[0].evalI(c)
		case text_align:
			ts.align = exp[0].evalI(c)
		case text_pos:
			ts.x = exp[0].evalF(c)/ts.localScale + float32(ts.offsetX)
			if len(exp) > 1 {
				ts.y = exp[1].evalF(c) / ts.localScale
			}
		case text_scale:
			xscl = exp[0].evalF(c)
			if len(exp) > 1 {
				yscl = exp[1].evalF(c)
			}
		case text_color:
			var r, g, b int32 = exp[0].evalI(c), 255, 255
			if len(exp) > 1 {
				g = exp[1].evalI(c)
				if len(exp) > 2 {
					b = exp[2].evalI(c)
				}
			}
			ts.SetColor(r, g, b)
		case text_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	ts.xscl = xscl / ts.localScale
	ts.yscl = yscl / ts.localScale
	if fnt == -1 {
		ts.fnt = sys.debugFont.fnt
		ts.xscl *= sys.debugFont.xscl
		ts.yscl *= sys.debugFont.yscl
	}
	if ts.text == "" {
		ts.text = OldSprintf("%v", params...)
	}
	sys.lifebar.textsprite = append(sys.lifebar.textsprite, ts)
	return false
}

// Platform bytecode definitons
type createPlatform StateControllerBase

const (
	createPlatform_id byte = iota
	createPlatform_name
	createPlatform_anim
	createPlatform_pos
	createPlatform_size
	createPlatform_offset
	createPlatform_activeTime
	createPlatform_redirectid
)

// The createPlatform bytecode function.
func (sc createPlatform) Run(schara *Char, _ []int32) bool {
	var chara = schara
	var customOffset = false
	var plat = Platform{
		anim:       -1,
		pos:        [2]float32{0, 0},
		size:       [2]int32{0, 0},
		offset:     [2]int32{0, 0},
		activeTime: -1,
	}

	StateControllerBase(sc).run(schara, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case createPlatform_id:
			plat.id = exp[0].evalI(schara)
		case createPlatform_name:
			plat.name = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		case createPlatform_pos:
			plat.pos[0] = exp[0].evalF(schara)
			plat.pos[1] = exp[1].evalF(schara)
		case createPlatform_size:
			plat.size[0] = exp[0].evalI(schara)
			plat.size[1] = exp[1].evalI(schara)
		case createPlatform_offset:
			customOffset = true
			plat.offset[0] = exp[0].evalI(schara)
			plat.offset[1] = exp[1].evalI(schara)
		case createPlatform_activeTime:
			plat.activeTime = exp[0].evalI(schara)
		case createPlatform_redirectid:
			if rid := sys.playerID(exp[0].evalI(schara)); rid != nil {
				chara = rid
			} else {
				return false
			}
		}
		return true
	})

	if !customOffset {
		if plat.size[0] != 0 {
			plat.offset[0] = plat.size[0] / 2
		}
		if plat.size[1] != 0 {
			plat.offset[1] = plat.size[1] / 2
		}
	}
	plat.ownerID = chara.id

	return false
}

type removePlatform StateControllerBase

const (
	removePlatform_id byte = iota
	removePlatform_name
)

type modifyStageVar StateControllerBase

const (
	modifyStageVar_camera_boundleft byte = iota
	modifyStageVar_camera_boundright
	modifyStageVar_camera_boundhigh
	modifyStageVar_camera_boundlow
	modifyStageVar_camera_verticalfollow
	modifyStageVar_camera_floortension
	modifyStageVar_camera_tensionhigh
	modifyStageVar_camera_tensionlow
	modifyStageVar_camera_tension
	modifyStageVar_camera_tensionvel
	modifyStageVar_camera_cuthigh
	modifyStageVar_camera_cutlow
	modifyStageVar_camera_startzoom
	modifyStageVar_camera_zoomout
	modifyStageVar_camera_zoomin
	modifyStageVar_camera_zoomindelay
	modifyStageVar_camera_ytension_enable
	modifyStageVar_camera_autocenter
	modifyStageVar_playerinfo_leftbound
	modifyStageVar_playerinfo_rightbound
	modifyStageVar_scaling_topscale
	modifyStageVar_bound_screenleft
	modifyStageVar_bound_screenright
	modifyStageVar_stageinfo_zoffset
	modifyStageVar_stageinfo_zoffsetlink
	modifyStageVar_stageinfo_xscale
	modifyStageVar_stageinfo_yscale
	modifyStageVar_shadow_intensity
	modifyStageVar_shadow_color
	modifyStageVar_shadow_yscale
	modifyStageVar_shadow_fade_range
	modifyStageVar_shadow_xshear
	modifyStageVar_reflection_intensity
	modifyStageVar_redirectid
)

func (sc modifyStageVar) Run(c *Char, _ []int32) bool {
	//crun := c
	s := *&sys.stage
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case modifyStageVar_camera_autocenter:
			s.stageCamera.autocenter = exp[0].evalB(c)
		case modifyStageVar_camera_boundleft:
			s.stageCamera.boundleft = exp[0].evalI(c)
		case modifyStageVar_camera_boundright:
			s.stageCamera.boundright = exp[0].evalI(c)
		case modifyStageVar_camera_boundhigh:
			s.stageCamera.boundhigh = exp[0].evalI(c)
		case modifyStageVar_camera_boundlow:
			s.stageCamera.boundlow = exp[0].evalI(c)
		case modifyStageVar_camera_verticalfollow:
			s.stageCamera.verticalfollow = exp[0].evalF(c)
		case modifyStageVar_camera_floortension:
			s.stageCamera.floortension = exp[0].evalI(c)
		case modifyStageVar_camera_tensionhigh:
			s.stageCamera.tensionhigh = exp[0].evalI(c)
		case modifyStageVar_camera_tensionlow:
			s.stageCamera.tensionlow = exp[0].evalI(c)
		case modifyStageVar_camera_tension:
			s.stageCamera.tension = exp[0].evalI(c)
		case modifyStageVar_camera_tensionvel:
			s.stageCamera.tensionvel = exp[0].evalF(c)
		case modifyStageVar_camera_cuthigh:
			s.stageCamera.cuthigh = exp[0].evalI(c)
		case modifyStageVar_camera_cutlow:
			s.stageCamera.cutlow = exp[0].evalI(c)
		case modifyStageVar_camera_startzoom:
			s.stageCamera.startzoom = exp[0].evalF(c)
		case modifyStageVar_camera_zoomout:
			s.stageCamera.zoomout = exp[0].evalF(c)
		case modifyStageVar_camera_zoomin:
			s.stageCamera.zoomin = exp[0].evalF(c)
		case modifyStageVar_camera_zoomindelay:
			s.stageCamera.zoomindelay = exp[0].evalF(c)
		case modifyStageVar_camera_ytension_enable:
			s.stageCamera.ytensionenable = exp[0].evalB(c)
		case modifyStageVar_playerinfo_leftbound:
			s.leftbound = exp[0].evalF(c)
		case modifyStageVar_playerinfo_rightbound:
			s.rightbound = exp[0].evalF(c)
		case modifyStageVar_scaling_topscale:
			if s.mugenver[0] != 1 { //mugen 1.0+ removed support for topscale
				s.stageCamera.ztopscale = exp[0].evalF(c)
			}
		case modifyStageVar_bound_screenleft:
			s.screenleft = exp[0].evalI(c)
		case modifyStageVar_bound_screenright:
			s.screenright = exp[0].evalI(c)
		case modifyStageVar_stageinfo_zoffset:
			s.stageCamera.zoffset = exp[0].evalI(c)
		case modifyStageVar_stageinfo_zoffsetlink:
			s.zoffsetlink = exp[0].evalI(c)
		case modifyStageVar_stageinfo_xscale:
			s.scale[0] = exp[0].evalF(c)
		case modifyStageVar_stageinfo_yscale:
			s.scale[1] = exp[0].evalF(c)
		case modifyStageVar_shadow_intensity:
			s.sdw.intensity = Clamp(exp[0].evalI(c), 0, 255)
		case modifyStageVar_shadow_color:
			// mugen 1.1 removed support for color
			if (s.mugenver[0] != 1 || s.mugenver[1] != 1) && (s.sff.header.Ver0 != 2 || s.sff.header.Ver2 != 1) {
				r := Clamp(exp[0].evalI(c), 0, 255)
				g := Clamp(exp[1].evalI(c), 0, 255)
				b := Clamp(exp[2].evalI(c), 0, 255)
				s.sdw.color = uint32(r<<16 | g<<8 | b)
			}
		case modifyStageVar_shadow_yscale:
			s.sdw.yscale = exp[0].evalF(c)
		case modifyStageVar_shadow_fade_range:
			s.sdw.fadeend = exp[0].evalI(c)
			s.sdw.fadebgn = exp[1].evalI(c)
		case modifyStageVar_shadow_xshear:
			s.sdw.xshear = exp[0].evalF(c)
		case modifyStageVar_reflection_intensity:
			s.reflection = Clamp(exp[0].evalI(c), 0, 255)
		case modifyStageVar_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				//crun = rid
			} else {
				return false
			}
		}
		return true
	})
	sys.stage.reload = true // Stage will have to be reloaded if it's re-selected
	sys.cam.stageCamera = s.stageCamera
	sys.cam.Reset()
	return false
}

type cameraCtrl StateControllerBase

const (
	cameraCtrl_view byte = iota
	cameraCtrl_pos
	cameraCtrl_followid
)

func (sc cameraCtrl) Run(c *Char, _ []int32) bool {
	//crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case cameraCtrl_view:
			sys.cam.View = CameraView(exp[0].evalI(c))
			if sys.cam.View == Follow_View {
				sys.cam.FollowChar = c
			}
		case cameraCtrl_pos:
			sys.cam.Pos[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				sys.cam.Pos[1] = exp[1].evalF(c)
			}
		case cameraCtrl_followid:
			if cid := sys.playerID(exp[0].evalI(c)); cid != nil {
				sys.cam.FollowChar = cid
			}
		}
		return true
	})
	return false
}

type height StateControllerBase

const (
	height_value byte = iota
	height_redirectid
)

func (sc height) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case height_value:
			crun.setTHeight(exp[0].evalF(c) * lclscround)
			if len(exp) > 1 {
				crun.setBHeight(exp[1].evalF(c) * lclscround)
			}
		case height_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = (320 / c.localcoord) / (320 / crun.localcoord)
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type modifyChar StateControllerBase

const (
	modifyChar_lifemax byte = iota
	modifyChar_powermax
	modifyChar_dizzypointsmax
	modifyChar_guardpointsmax
	modifyChar_teamside
	modifyChar_displayname
	modifyChar_lifebarname
	modifyChar_redirectid
)

func (sc modifyChar) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case modifyChar_lifemax:
			lm := exp[0].evalI(c)
			if lm < 1 {
				lm = 1
			}
			crun.lifeMax = lm
			crun.life = Clamp(crun.life, 0, crun.lifeMax)
		case modifyChar_powermax:
			pm := exp[0].evalI(c)
			if pm < 0 {
				pm = 0
			}
			crun.powerMax = pm
			crun.power = Clamp(crun.power, 0, crun.powerMax)
		case modifyChar_dizzypointsmax:
			dp := exp[0].evalI(c)
			if dp < 0 {
				dp = 0
			}
			crun.dizzyPointsMax = dp
			crun.dizzyPoints = Clamp(crun.dizzyPoints, 0, crun.dizzyPointsMax)
		case modifyChar_guardpointsmax:
			gp := exp[0].evalI(c)
			if gp < 0 {
				gp = 0
			}
			crun.guardPointsMax = gp
			crun.guardPoints = Clamp(crun.guardPoints, 0, crun.guardPointsMax)
		case modifyChar_teamside:
			ts := int(exp[0].evalI(c))
			if ts >= 0 && ts <= 2 {
				ts -= 1 // Internally the teamside goes from -1 to 1
				crun.teamside = ts
			}
		case modifyChar_displayname:
			dn := string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			sys.cgi[crun.playerNo].displayname = dn
		case modifyChar_lifebarname:
			ln := string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			sys.cgi[crun.playerNo].lifebarname = ln
		case modifyChar_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type getHitVarSet StateControllerBase

const (
	getHitVarSet_airtype byte = iota
	getHitVarSet_animtype
	getHitVarSet_attr
	getHitVarSet_chainid
	getHitVarSet_ctrltime
	getHitVarSet_fall
	getHitVarSet_fall_damage
	getHitVarSet_fall_envshake_ampl
	getHitVarSet_fall_envshake_freq
	getHitVarSet_fall_envshake_mul
	getHitVarSet_fall_envshake_phase
	getHitVarSet_fall_envshake_time
	getHitVarSet_fall_kill
	getHitVarSet_fall_recover
	getHitVarSet_fall_recovertime
	getHitVarSet_fall_xvel
	getHitVarSet_fall_yvel
	getHitVarSet_fallcount
	getHitVarSet_ground_animtype
	getHitVarSet_groundtype
	getHitVarSet_guarded
	getHitVarSet_hitshaketime
	getHitVarSet_hittime
	getHitVarSet_id
	getHitVarSet_playerno
	getHitVarSet_recovertime
	getHitVarSet_slidetime
	getHitVarSet_xvel
	getHitVarSet_yaccel
	getHitVarSet_yvel
	getHitVarSet_redirectid
)

func (sc getHitVarSet) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case getHitVarSet_airtype:
			crun.ghv.airtype = HitType(exp[0].evalI(c))
		case getHitVarSet_animtype:
			crun.ghv.animtype = Reaction(exp[0].evalI(c))
		case getHitVarSet_attr:
			crun.ghv.attr = exp[0].evalI(c)
		case getHitVarSet_chainid:
			crun.ghv.hitid = exp[0].evalI(c)
		case getHitVarSet_ctrltime:
			crun.ghv.ctrltime = exp[0].evalI(c)
		case getHitVarSet_fall:
			crun.ghv.fallf = exp[0].evalB(c)
		case getHitVarSet_fall_damage:
			crun.ghv.fall.damage = exp[0].evalI(c)
		case getHitVarSet_fall_envshake_ampl:
			crun.ghv.fall.envshake_ampl = int32(exp[0].evalF(c) * lclscround)
		case getHitVarSet_fall_envshake_freq:
			crun.ghv.fall.envshake_freq = exp[0].evalF(c)
		case getHitVarSet_fall_envshake_mul:
			crun.ghv.fall.envshake_mul = exp[0].evalF(c)
		case getHitVarSet_fall_envshake_phase:
			crun.ghv.fall.envshake_phase = exp[0].evalF(c)
		case getHitVarSet_fall_envshake_time:
			crun.ghv.fall.envshake_time = exp[0].evalI(c)
		case getHitVarSet_fall_kill:
			crun.ghv.fall.kill = exp[0].evalB(c)
		case getHitVarSet_fall_recover:
			crun.ghv.fall.recover = exp[0].evalB(c)
		case getHitVarSet_fall_recovertime:
			crun.ghv.fall.recovertime = exp[0].evalI(c)
		case getHitVarSet_fall_xvel:
			crun.ghv.fall.xvelocity = exp[0].evalF(c) * lclscround
		case getHitVarSet_fall_yvel:
			crun.ghv.fall.yvelocity = exp[0].evalF(c) * lclscround
		case getHitVarSet_fallcount:
			crun.ghv.fallcount = exp[0].evalI(c)
		case getHitVarSet_groundtype:
			crun.ghv.groundtype = HitType(exp[0].evalI(c))
		case getHitVarSet_guarded:
			crun.ghv.guarded = exp[0].evalB(c)
		case getHitVarSet_hittime:
			crun.ghv.hittime = exp[0].evalI(c)
		case getHitVarSet_hitshaketime:
			crun.ghv.hitshaketime = exp[0].evalI(c)
		case getHitVarSet_id:
			crun.ghv.id = exp[0].evalI(c)
		case getHitVarSet_playerno:
			crun.ghv.playerNo = int(exp[0].evalI(c))
		case getHitVarSet_recovertime:
			crun.recoverTime = exp[0].evalI(c)
		case getHitVarSet_slidetime:
			crun.ghv.slidetime = exp[0].evalI(c)
		case getHitVarSet_xvel:
			crun.ghv.xvel = exp[0].evalF(c) * crun.facing * lclscround
		case getHitVarSet_yaccel:
			crun.ghv.yaccel = exp[0].evalF(c) * lclscround
		case getHitVarSet_yvel:
			crun.ghv.yvel = exp[0].evalF(c) * lclscround
		case getHitVarSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type groundLevelOffset StateControllerBase

const (
	groundLevelOffset_value byte = iota
	groundLevelOffset_redirectid
)

func (sc groundLevelOffset) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case groundLevelOffset_value:
			crun.groundLevel = exp[0].evalF(c) * lclscround
		case groundLevelOffset_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
			} else {
				return false
			}
		}
		return true
	})
	return false
}

// StateDef data struct
type StateBytecode struct {
	stateType StateType
	moveType  MoveType
	physics   StateType
	playerNo  int
	stateDef  stateDef
	block     StateBlock
	ctrlsps   []int32
	numVars   int32
}

// StateDef bytecode creation function
func newStateBytecode(pn int) *StateBytecode {
	sb := &StateBytecode{
		stateType: ST_S,
		moveType:  MT_I,
		physics:   ST_N,
		playerNo:  pn,
		block:     *newStateBlock(),
	}
	return sb
}
func (sb *StateBytecode) init(c *Char) {
	if sb.stateType != ST_U {
		c.ss.changeStateType(sb.stateType)
	}
	if sb.moveType != MT_U {
		if !c.ss.storeMoveType {
			c.ss.prevMoveType = c.ss.moveType
		}
		c.ss.moveType = sb.moveType
	}
	if sb.physics != ST_U {
		c.ss.physics = sb.physics
	}
	c.ss.storeMoveType = false
	sys.workingState = sb
	sb.stateDef.Run(c)
}
func (sb *StateBytecode) run(c *Char) (changeState bool) {
	sys.bcVar = sys.bcVarStack.Alloc(int(sb.numVars))
	sys.workingState = sb
	changeState = sb.block.Run(c, sb.ctrlsps)
	if len(sys.bcStack) != 0 {
		sys.errLog.Println(sys.cgi[sb.playerNo].def)
		for _, v := range sys.bcStack {
			sys.errLog.Printf("%+v\n", v)
		}
		c.panic()
	}
	sys.bcVarStack.Clear()
	return
}
