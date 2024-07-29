package main

import "math"

type stageCamera struct {
	startx                  int32
	starty                  int32
	boundleft               int32
	boundright              int32
	boundhigh               int32
	boundlow                int32
	verticalfollow          float32
	floortension            int32
	tensionhigh             int32
	tensionlow              int32
	lowestcap               bool
	tension                 int32
	tensionvel              float32
	overdrawhigh            int32 // TODO: not implemented
	overdrawlow             int32
	cuthigh                 int32
	cutlow                  int32
	localcoord              [2]int32
	localscl                float32
	zoffset                 int32
	ztopscale               float32
	startzoom               float32
	zoomin                  float32
	zoomout                 float32
	ytensionenable          bool
	autocenter              bool
	zoomanchor              bool
	boundhighzoomdelta      float32
	verticalfollowzoomdelta float32
	zoomindelay             float32
	zoomindelaytime         float32
	zoominspeed             float32
	zoomoutspeed            float32
	yscrollspeed            float32
	fov                     float32
	yshift                  float32
	far                     float32
	near                    float32
	aspectcorrection        float32
	zoomanchorcorrection    float32
	ywithoutbound           float32
	highest                 float32
	lowest                  float32
	leftest                 float32
	rightest                float32
	leftestvel              float32
	rightestvel             float32
	roundstart              bool
	maxRight                float32
	minLeft                 float32
}

func newStageCamera() *stageCamera {
	return &stageCamera{verticalfollow: 0.2, tensionvel: 1, tension: 50,
		cuthigh: 0, cutlow: math.MinInt32,
		localcoord: [...]int32{320, 240}, localscl: float32(sys.gameWidth / 320),
		ztopscale: 1, startzoom: 1, zoomin: 1, zoomout: 1, ytensionenable: false,
		tensionhigh: 0, tensionlow: 0,
		fov: 40, yshift: 0, far: 10000, near: 0.1,
		zoomindelay: 0, zoominspeed: 1, zoomoutspeed: 1, yscrollspeed: 1,
		boundhighzoomdelta: 0, verticalfollowzoomdelta: 0}
}

type CameraView int

const (
	Fighting_View CameraView = iota
	Follow_View
	Free_View
)

type Camera struct {
	stageCamera
	View                            CameraView
	ZoomEnable, ZoomActive          bool
	ZoomDelayEnable                 bool
	ZoomMin, ZoomMax, ZoomSpeed     float32
	zoomdelay                       float32
	Pos, ScreenPos, Offset          [2]float32
	XMin, XMax                      float32
	Scale, MinScale                 float32
	boundL, boundR, boundH, boundLo float32
	zoff                            float32
	halfWidth                       float32
	FollowChar                      *Char
}

func newCamera() *Camera {
	return &Camera{View: Fighting_View, ZoomMin: 5.0 / 6, ZoomMax: 15.0 / 14, ZoomSpeed: 12}
}
func (c *Camera) Reset() {
	c.ZoomEnable = c.ZoomActive && (c.stageCamera.zoomin != 1 || c.stageCamera.zoomout != 1)
	c.boundL = float32(c.boundleft-c.startx)*c.localscl - ((1-c.zoomout)*100*c.zoomout)*(1/c.zoomout)*(1/c.zoomout)*1.6*(float32(sys.gameWidth)/320)
	c.boundR = float32(c.boundright-c.startx)*c.localscl + ((1-c.zoomout)*100*c.zoomout)*(1/c.zoomout)*(1/c.zoomout)*1.6*(float32(sys.gameWidth)/320)
	c.halfWidth = float32(sys.gameWidth) / 2
	c.XMin = c.boundL - c.halfWidth/c.BaseScale()
	c.XMax = c.boundR + c.halfWidth/c.BaseScale()
	c.aspectcorrection = 0
	c.zoomanchorcorrection = 0
	c.zoomin = MaxF(c.zoomin, c.zoomout)
	if c.cutlow == math.MinInt32 {
		c.cutlow = int32(float32(c.localcoord[1]-c.zoffset) - float32(c.localcoord[1])*0.05)
	}
	if float32(c.localcoord[1])*c.localscl-float32(sys.gameHeight) < 0 {
		c.aspectcorrection = MinF(0, (float32(c.localcoord[1])*c.localscl-float32(sys.gameHeight))+MinF((float32(sys.gameHeight)-float32(c.localcoord[1])*c.localscl)/2, float32(c.overdrawlow)*c.localscl))
	} else if float32(c.localcoord[1])*c.localscl-float32(sys.gameHeight) > 0 {
		if c.cuthigh+c.cutlow <= 0 {
			c.aspectcorrection = float32(Ceil(float32(c.localcoord[1])*c.localscl) - sys.gameHeight)
		} else {
			diff := Ceil(float32(c.localcoord[1])*c.localscl) - sys.gameHeight
			tmp := Ceil(float32(c.cuthigh)*c.localscl) * diff / (Ceil(float32(c.cuthigh)*c.localscl) + Ceil(float32(c.cutlow)*c.localscl))
			if diff-tmp <= c.cutlow {
				c.aspectcorrection = float32(tmp)
			} else {
				c.aspectcorrection = float32(diff - Ceil(float32(c.cutlow)*c.localscl))
			}
		}

	}
	c.boundH = float32(c.boundhigh) * c.localscl
	c.boundLo = float32(Max(c.boundhigh, c.boundlow)) * c.localscl
	c.boundlow = Max(c.boundhigh, c.boundlow)
	c.tensionvel = MaxF(MinF(c.tensionvel, 20), 0)
	if c.verticalfollow < 0 {
		c.ytensionenable = true
	}
	xminscl := float32(sys.gameWidth) / (float32(sys.gameWidth) - c.boundL +
		c.boundR)
	//yminscl := float32(sys.gameHeight) / (240 - MinF(0, c.boundH))
	c.MinScale = MaxF(c.zoomout, MinF(c.zoomin, xminscl))
	c.maxRight = float32(c.boundright)*c.localscl + c.halfWidth/c.zoomout
	c.minLeft = float32(c.boundleft)*c.localscl - c.halfWidth/c.zoomout
}
func (c *Camera) Init() {
	c.Reset()
	c.View = Fighting_View
	c.roundstart = true
	c.Scale = c.startzoom
	c.Pos[0], c.Pos[1], c.ywithoutbound = float32(c.startx)*c.localscl, float32(c.starty)*c.localscl, float32(c.starty)*c.localscl
	c.zoomindelaytime = c.zoomindelay
}
func (c *Camera) ResetTracking() {
	c.leftest = c.Pos[0]
	c.rightest = c.Pos[0]
	c.highest = math.MaxFloat32
	c.lowest = -math.MaxFloat32
	c.leftestvel = 0
	c.rightestvel = 0
}
func (c *Camera) Update(scl, x, y float32) {
	c.Scale = c.BaseScale() * scl
	c.zoff = float32(c.zoffset) * c.localscl
	if sys.stage.stageCamera.zoomanchor {
		c.zoomanchorcorrection = c.zoff - (float32(sys.gameHeight) + c.aspectcorrection - (float32(sys.gameHeight)-c.zoff+c.aspectcorrection)*scl)
	}
	for i := 0; i < 2; i++ {
		c.Offset[i] = sys.stage.bga.offset[i] * sys.stage.localscl * scl
	}
	c.ScreenPos[0] = x - c.halfWidth/c.Scale - c.Offset[0]
	c.ScreenPos[1] = y - (c.GroundLevel()-float32(sys.gameHeight-240)*scl)/
		c.Scale - c.Offset[1]
	c.Pos[0] = x
	c.Pos[1] = y
}
func (c *Camera) ScaleBound(scl, sclmul float32) float32 {
	if c.ZoomEnable {
		if sys.debugPaused() {
			sclmul = 1
		} else if sys.turbo < 1 {
			sclmul = Pow(sclmul, sys.turbo)
		}
		return MaxF(c.MinScale, MinF(c.zoomin, scl*sclmul))
	}
	return 1
}
func (c *Camera) XBound(scl, x float32) float32 {
	return ClampF(x,
		c.boundL-c.halfWidth+c.halfWidth/scl,
		c.boundR+c.halfWidth-c.halfWidth/scl)
}
func (c *Camera) BaseScale() float32 {
	return c.ztopscale
}
func (c *Camera) GroundLevel() float32 {
	return c.zoff - c.aspectcorrection - c.zoomanchorcorrection
}
func (c *Camera) ResetZoomdelay() {
	c.zoomdelay = 0
}
func (c *Camera) action(x, y, scale float32, pause bool) (newX, newY, newScale float32) {
	newX = x
	newY = y
	newScale = scale
	if !sys.debugPaused() {
		newY = y / scale
		switch c.View {
		case Fighting_View:
			if c.highest != math.MaxFloat32 && c.lowest != -math.MaxFloat32 {
				if c.lowestcap {
					c.lowest = MaxF(c.lowest, float32(c.boundhigh)*c.localscl-(float32(sys.gameHeight)-c.GroundLevel()-float32(c.tensionlow))/c.zoomout)
				}
				tension := MaxF(0, float32(c.tension)*c.localscl)
				oldLeft, oldRight := x-c.halfWidth/scale, x+c.halfWidth/scale
				targetLeft, targetRight := oldLeft, oldRight
				if c.autocenter {
					targetLeft = MinF(MaxF((c.leftest+c.rightest)/2-c.halfWidth/scale, c.minLeft), c.maxRight-2*c.halfWidth/scale)
					targetRight = targetLeft + 2*c.halfWidth/scale
				}

				if c.leftest < targetLeft+tension {
					diff := targetLeft - MaxF(c.leftest-tension, c.minLeft)
					targetLeft = MaxF(c.leftest-tension, c.minLeft)
					targetRight = MaxF(oldRight-diff, MinF(c.rightest+tension, c.maxRight))
				} else if c.rightest > targetRight-tension {
					diff := targetRight - MinF(c.rightest+tension, c.maxRight)
					targetRight = MinF(c.rightest+tension, c.maxRight)
					targetLeft = MinF(oldLeft-diff, MaxF(c.leftest-tension, c.minLeft))
				}
				if c.halfWidth*2/(targetRight-targetLeft) < c.zoomout {
					rLeft, rRight := MaxF(targetLeft+tension-c.leftest, 0), MaxF(c.rightest-(targetRight-tension), 0)
					diff := 2 * ((targetRight-targetLeft)/2 - c.halfWidth/c.zoomout)
					if rLeft > rRight {
						diff2 := rLeft - rRight
						targetRight -= MinF(diff2, diff)
						diff -= MinF(diff2, diff)
					} else if rRight > rLeft {
						diff2 := rRight - rLeft
						targetLeft += MinF(diff2, diff)
						diff -= MinF(diff2, diff)
					}
					targetLeft += diff / 2
					targetRight -= diff / 2
					if c.leftest-targetLeft < float32(sys.stage.screenleft)*c.localscl {
						diff := MinF(float32(sys.stage.screenleft)*c.localscl-(c.leftest-targetLeft), targetLeft-c.minLeft)
						if targetRight-c.rightest < float32(sys.stage.screenright)*c.localscl {
							diff2 := MinF(float32(sys.stage.screenright)*c.localscl-(targetRight-c.rightest), c.maxRight-targetRight)
							//diff = diff + (MinF(float32(sys.stage.screenright)*c.localscl-(targetRight-c.rightest), c.maxRight-targetRight)-diff)/2
							diff = diff - diff2
						}
						targetLeft -= diff
						targetRight -= diff
					} else if targetRight-c.rightest < float32(sys.stage.screenright)*c.localscl {
						diff := MinF(float32(sys.stage.screenright)*c.localscl-(targetRight-c.rightest), c.maxRight-targetRight)
						targetLeft += diff
						targetRight += diff
					}
				}
				maxScale := c.zoomin
				if c.ytensionenable {
					maxScale = MinF(MaxF(float32(sys.gameHeight)/((c.lowest+float32(c.tensionlow)*c.localscl)-(c.highest-float32(c.tensionhigh)*c.localscl)), c.zoomout), maxScale)
				}
				if c.halfWidth*2/(targetRight-targetLeft) < maxScale {
					if c.zoomindelaytime > 0 {
						c.zoomindelaytime -= 1
					} else {
						diffLeft := MaxF(c.leftest-tension-targetLeft, 0)
						if diffLeft < 0 {
							diffLeft = 0
						}
						diffRight := MinF(c.rightest+tension-targetRight, 0)
						if diffRight > 0 {
							diffRight = 0
						}
						if c.halfWidth*2/((targetRight+diffRight)-(targetLeft+diffLeft)) > maxScale {
							tmp, tmp2 := diffLeft/(diffLeft-diffRight)*((targetRight+diffRight)-(targetLeft+diffLeft)-c.halfWidth*2/maxScale), diffRight/(diffLeft-diffRight)*((targetRight+diffRight)-(targetLeft+diffLeft)-c.halfWidth*2/maxScale)
							diffLeft += tmp
							diffRight += tmp2
						}
						if c.halfWidth*2/((targetRight+diffRight)-(targetLeft+diffLeft)) > scale {
							targetLeft += diffLeft
							targetRight += diffRight
						} else {
							c.zoomindelaytime = c.zoomindelay
						}
					}
				} else {
					c.zoomindelaytime = c.zoomindelay
				}

				targetX := (targetLeft + targetRight) / 2
				targetScale := MinF(c.halfWidth*2/(targetRight-targetLeft), maxScale)

				if !c.ytensionenable {
					//newY = c.ywithoutbound
					ywithoutbound := c.ywithoutbound
					verticalfollow := MaxF(c.verticalfollow, 0.0) + (targetScale-c.zoomout)*MaxF(c.verticalfollowzoomdelta, 0.0)
					targetY := (c.highest + float32(c.floortension)*c.localscl) * verticalfollow
					if !c.roundstart {
						for i := 0; i < 3; i++ {
							ywithoutbound = ywithoutbound*.85 + targetY*.15
							if AbsF(targetY-ywithoutbound)*sys.heightScale < 1 {
								ywithoutbound = targetY
							}
							if AbsF(newY-ywithoutbound) < float32(sys.gameWidth)/320*5.5 {
								newY = ywithoutbound
							} else {
								if newY > ywithoutbound {
									newY -= float32(sys.gameWidth) / 320 * 0.5
									newY -= (newY - ywithoutbound) * verticalfollow / 10
								} else {
									newY += float32(sys.gameWidth) / 320 * 0.5
									newY += (ywithoutbound - newY) * verticalfollow / 10
								}
							}
						}
					} else {
						ywithoutbound = targetY
						newY = ywithoutbound
					}
					c.ywithoutbound = ywithoutbound
				} else {
					targetScale = MinF(MinF(MaxF(float32(sys.gameHeight)/((c.lowest+float32(c.tensionlow)*c.localscl)-(c.highest-float32(c.tensionhigh)*c.localscl)), c.zoomout), c.zoomin), targetScale)
					targetX = MinF(MaxF(targetX, float32(c.boundleft)*c.localscl-c.halfWidth*(1/c.zoomout-1/targetScale)), float32(c.boundright)*c.localscl+c.halfWidth*(1/c.zoomout-1/targetScale))
					targetLeft = targetX - c.halfWidth/targetScale
					targetRight = targetX + c.halfWidth/targetScale

					newY = c.ywithoutbound
					targetY := c.GroundLevel()/targetScale + (c.highest - float32(c.tensionhigh)*c.localscl)
					if !c.roundstart {
						diff := float32(sys.gameWidth) / 320 * 2.5
						for i := 0; i < 3; i++ {
							newY = (newY + targetY) * .5
							if AbsF(targetY-newY) < diff {
								newY = targetY
								break
							} else if targetY-newY > diff {
								newY = newY + diff
							} else {
								newY = newY - diff
							}
						}
					} else {
						newY = targetY
					}
					c.ywithoutbound = newY
				}

				newLeft, newRight := oldLeft, oldRight
				if !c.roundstart {
					diff := float32(sys.gameWidth) / 3200
					for i := 0; i < 3; i++ {
						newLeft, newRight = newLeft+(targetLeft-newLeft)*0.05*sys.turbo*c.tensionvel, newRight+(targetRight-newRight)*0.05*sys.turbo*c.tensionvel
						diffLeft := targetLeft - newLeft
						diffRight := targetRight - newRight
						if AbsF(diffLeft) <= diff*sys.turbo*c.tensionvel {
							newLeft = targetLeft
						} else if diffLeft > 0 {
							newLeft += diff * sys.turbo * c.tensionvel
						} else {
							newLeft -= diff * sys.turbo * c.tensionvel
						}
						if newLeft-oldLeft > 0 && newLeft-oldLeft < c.rightestvel {
							newLeft = MinF(oldLeft+c.rightestvel, targetLeft)
						} else if newLeft-oldLeft < 0 && newLeft-oldLeft > c.leftestvel {
							newLeft = MaxF(oldLeft+c.leftestvel, targetLeft)
						}

						if AbsF(diffRight) <= diff*sys.turbo*c.tensionvel {
							newRight = targetRight
						} else if diffRight > 0 {
							newRight += diff * sys.turbo * c.tensionvel
						} else {
							newRight -= diff * sys.turbo * c.tensionvel
						}
						if newRight-oldRight > 0 && newRight-oldRight < c.rightestvel {
							newRight = MinF(oldRight+c.rightestvel, targetRight)
						} else if newRight-oldRight < 0 && newRight-oldRight > c.leftestvel {
							newRight = MaxF(oldRight+c.leftestvel, targetRight)
						}
					}
				} else {
					newLeft, newRight = targetLeft, targetRight
				}
				newScale = MinF(c.halfWidth*2/(newRight-newLeft), c.zoomin)
				newLeft, newRight, newScale = c.reduceZoomSpeed(newLeft, newRight, newScale, oldLeft, oldRight, scale)
				newX = (newLeft + newRight) / 2
				newY = c.reduceYScrollSpeed(newY, y)
				newY = c.boundY(newY, newScale)
			} else {
				newScale = MinF(MaxF(newScale, c.zoomout), c.zoomin)
				newX = MinF(MaxF(newX, c.minLeft+c.halfWidth/newScale), c.maxRight-c.halfWidth/newScale)
				newY = c.boundY(newY, newScale)
			}

		case Follow_View:
			newX = c.FollowChar.pos[0]
			newY = c.FollowChar.pos[1] * Pow(c.verticalfollow, MinF(1, 1/Pow(c.Scale, 4)))
			newScale = 1
		case Free_View:
			newX = c.Pos[0]
			newY = c.Pos[1]
			c.ywithoutbound = newY
			newScale = 1
		}
	}
	c.roundstart = false
	return
}

func (c *Camera) reduceZoomSpeed(newLeft float32, newRight float32, newScale float32, oldLeft float32, oldRight float32, oldScale float32) (float32, float32, float32) {
	const minBoundDiff float32 = 5e-5
	const minScaleDiff float32 = 5e-4

	var speedFactor float32
	if newScale > oldScale {
		speedFactor = c.zoominspeed
	} else {
		speedFactor = c.zoomoutspeed
	}

	if speedFactor < 0.0 || speedFactor >= 1.0 {
		return newLeft, newRight, newScale
	}

	scaleDiff := newScale - oldScale
	leftAbsDiff, rightAbsDiff := AbsF(newLeft-oldLeft), AbsF(newRight-oldRight)

	if AbsF(scaleDiff) < minScaleDiff || (leftAbsDiff < minBoundDiff && rightAbsDiff < minBoundDiff) {
		return newLeft, newRight, newScale
	}

	adjustedNewScale := oldScale + speedFactor*scaleDiff
	scaleAdjustmentFactor := adjustedNewScale / newScale

	width := newRight - newLeft
	widthAdjustmentFactor := 1.0 / scaleAdjustmentFactor
	widthAdjustmentDiff := width*widthAdjustmentFactor - width

	totalAbsDiff := leftAbsDiff + rightAbsDiff
	adjustedNewLeft := newLeft - widthAdjustmentDiff*leftAbsDiff/totalAbsDiff
	adjustedNewRight := newRight + widthAdjustmentDiff*rightAbsDiff/totalAbsDiff

	adjustedNewLeft, adjustedNewRight = c.keepScreenEdge(adjustedNewLeft, adjustedNewRight)
	adjustedNewLeft, adjustedNewRight = c.keepStageEdge(adjustedNewLeft, adjustedNewRight)

	return c.hardLimit(adjustedNewLeft, adjustedNewRight)
}

func (c *Camera) keepScreenEdge(left float32, right float32) (float32, float32) {
	screenLeftest := c.leftest - float32(sys.stage.screenleft)*c.localscl
	if left > screenLeftest {
		right += screenLeftest - left
		left = screenLeftest
	}

	screenRightest := c.rightest + float32(sys.stage.screenright)*c.localscl
	if right < screenRightest {
		left += screenRightest - right
		right = screenRightest
	}

	return left, right
}

func (c *Camera) keepStageEdge(left float32, right float32) (float32, float32) {
	if left < c.minLeft {
		right += c.minLeft - left
		left = c.minLeft
	}
	if right > c.maxRight {
		left += c.maxRight - right
		right = c.maxRight
	}
	return left, right
}

func (c *Camera) hardLimit(left float32, right float32) (float32, float32, float32) {
	left = MaxF(left, c.minLeft)
	right = MinF(right, c.maxRight)
	scale := MaxF(MinF(c.halfWidth*2/(right-left), c.zoomin), c.zoomout)
	return left, right, scale
}

func (c *Camera) reduceYScrollSpeed(newY float32, oldY float32) float32 {
	const minYDiff float32 = 5e-5

	yDiff := newY - oldY
	if AbsF(yDiff) < minYDiff || c.yscrollspeed < 0.0 || c.yscrollspeed >= 1.0 {
		return newY
	}

	return oldY + yDiff*c.yscrollspeed
}

func (c *Camera) boundY(y float32, scale float32) float32 {
	if c.boundhighzoomdelta > 0 {
		topBound := float32(c.boundhigh)*c.localscl - c.GroundLevel()/c.zoomout
		boundHigh := float32(c.boundhigh)*c.localscl + ((topBound+c.GroundLevel()/scale)-float32(c.boundhigh)*c.localscl)/c.boundhighzoomdelta
		return MinF(MaxF(y, boundHigh), float32(c.boundlow)*c.localscl) * scale
	} else {
		return MinF(MaxF(y, float32(c.boundhigh)*c.localscl), float32(c.boundlow)*c.localscl) * scale
	}
}
