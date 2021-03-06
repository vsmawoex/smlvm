package arch

// VirtMemory defines an address space which page mapping is defined by a
// page table.  If no page table is defined, direct mapping is used.
type virtMemory struct {
	phyMem *phyMemory
	ptable *pageTable
}

// NewVirtMemory creates a new virtual address space with no page table.
func newVirtMemory(phy *phyMemory) *virtMemory {
	ret := new(virtMemory)
	ret.phyMem = phy
	return ret
}

// SetTable applies a particular pagetable at a physical memory position.
// If the address is not page size aligned, it will be aligned down.
// If the address is 0, it will use direct mapping.
func (vm *virtMemory) SetTable(root uint32) {
	if root == 0 {
		vm.ptable = nil
	} else {
		vm.ptable = newPageTable(vm.phyMem, root)
	}
}

func (vm *virtMemory) transRead(addr uint32, ring byte) (uint32, *Excep) {
	if vm.ptable == nil {
		return addr, nil
	}
	return vm.ptable.TranslateRead(addr, ring)
}

func (vm *virtMemory) transWrite(addr uint32, ring byte) (uint32, *Excep) {
	if vm.ptable == nil {
		return addr, nil
	}
	return vm.ptable.TranslateWrite(addr, ring)
}

// ReadU32 reads the byte at the given virtual address.
func (vm *virtMemory) ReadU32(addr uint32, ring byte) (uint32, *Excep) {
	addr, e := vm.transRead(addr, ring)
	if e != nil {
		return 0, e
	}
	return vm.phyMem.ReadU32(addr)
}

// WriteU32 writes the byte at the given virtual address.
func (vm *virtMemory) WriteU32(addr uint32, ring byte, v uint32) *Excep {
	addr, e := vm.transWrite(addr, ring)
	if e != nil {
		return e
	}
	return vm.phyMem.WriteU32(addr, v)
}

// ReadU8 reads the byte at the given virtual address.
func (vm *virtMemory) ReadU8(addr uint32, ring byte) (byte, *Excep) {
	addr, e := vm.transRead(addr, ring)
	if e != nil {
		return 0, e
	}
	return vm.phyMem.ReadU8(addr)
}

// WriteU8 writes a byte at the given virtual address under
// a certain ring.
func (vm *virtMemory) WriteU8(addr uint32, ring byte, v byte) *Excep {
	addr, e := vm.transWrite(addr, ring)
	if e != nil {
		return e
	}
	return vm.phyMem.WriteU8(addr, v)
}
