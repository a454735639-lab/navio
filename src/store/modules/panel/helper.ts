import { ss } from '@/utils/storage'
import { PanelPanelConfigStyleEnum, PanelStateNetworkModeEnum } from '@/enums'
import defaultBackground from '@/assets/defaultBackground.jpg'
const LOCAL_NAME = 'panelStorage'

const defaultFooterHtml = ''

export function defaultStatePanelConfig(): Panel.panelConfig {
  return {
    backgroundImageSrc: defaultBackground,
    backgroundBlur: 0,
    backgroundMaskNumber: 0,
    iconStyle: PanelPanelConfigStyleEnum.info,
    iconTextColor: '#ffffff',
    iconTextInfoHideDescription: false,
    iconTextIconHideTitle: false,
    logoShow: true,
    logoText: 'Navio',
    logoImageSrc: '',
    clockShow: true,
    clockShowSecond: false,
    searchBoxShow: false,
    searchBoxSearchIcon: false,
    searchBoxBackground: '#2a2a2a6b',
    searchBoxBorderColor: '#cccccc',
    searchBoxTextColor: '#ffffff',
    marginBottom: 5,
    marginTop: 5,
    maxWidth: 1200,
    maxWidthUnit: 'px',
    marginX: 5,
    footerHtml: defaultFooterHtml,
    systemMonitorShow: false,
    systemMonitorShowTitle: true,
    systemMonitorPublicVisitModeShow: false,
    netModeChangeButtonShow: true,
    netModeChangeAuto: false,
    netModeChangeAutoMaxPingMs: 600,
    cardAliveDotShow: false,
    netProbeLanUrl: '',
    netProbeTsUrl: '',
  }
}

export function defaultState(): Panel.State {
  return {
    rightSiderCollapsed: false,
    leftSiderCollapsed: false,
    networkMode: PanelStateNetworkModeEnum.wan,
    panelConfig: { ...defaultStatePanelConfig() },
    itemGroupCardCacheData: [],
  }
}

export function getLocalState(): Panel.State {
  const localState = ss.get(LOCAL_NAME)
  return { ...defaultState(), ...localState }
}

export function setLocalState(state: Panel.State) {
  ss.set(LOCAL_NAME, state)
}

export function removeLocalState() {
  ss.remove(LOCAL_NAME)
}
