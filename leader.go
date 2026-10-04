package main

// ============================================================
// ПРАВИЛА ЛИДЕРА
// ============================================================
//
//  1. Лидера не бросают: при эвакуации его выносят первым и без риска потерять.
//  2. Храм ставит упавшего лидера на ноги бесплатно, а Гильдия никогда его не заменяет;
//     от алтаря и ловушек лидер не погибает, а лишь теряет сознание.
//  3. Падение лидера в бою пугает остальных (leaderShockStress).
//  4. Пока лидер на ногах, действует его призвание (см. LeaderCalling).

const leaderShockStress = 14

// leader возвращает героя-лидера отряда (nil, если лидера нет).
func (m *Model) leader() *Hero {
	for _, h := range m.Party {
		if h.IsLeader {
			return h
		}
	}
	return nil
}

// activeCalling возвращает призвание лидера, если он сейчас на ногах.
func (m *Model) activeCalling() LeaderCalling {
	l := m.leader()
	if l == nil || l.IsDead || l.IsDowned {
		return ""
	}
	return l.Calling
}

// onLeaderDowned вызывается, когда герой упал без сознания: падение лидера бьёт по духу отряда.
func (m *Model) onLeaderDowned(h *Hero) {
	if !h.IsLeader {
		return
	}
	m.addLog(dangerStyle.Render(T(m.Lang, "leader.shock", h.DisplayName(m.Lang), leaderShockStress)))
	for _, ally := range m.Party {
		if ally != h && !ally.IsDead && !ally.IsDowned {
			m.addStress(ally, leaderShockStress)
		}
	}
}

// hazardDeath обрабатывает гибель героя от алтаря или ловушки: обычный герой погибает
// навсегда, а лидер лишь теряет сознание и ждёт помощи Храма.
func (m *Model) hazardDeath(h *Hero) {
	if !h.IsLeader {
		m.recordFallenHero(h)
		return
	}
	h.HP = 0
	h.IsDowned = true
	h.IsGuarding = false
	h.IsBerserk = false
	h.IsStealthed = false
	h.IsCharged = false
	h.IsAura = false
	m.onLeaderDowned(h)
}
