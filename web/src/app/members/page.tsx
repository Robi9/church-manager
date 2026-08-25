"use client";

import { useEffect, useState, useCallback } from "react";
import Link from "next/link";
import { useAuth } from "@/lib/auth";
import { api } from "@/lib/api";
import { formatBrazilianPhone } from "@/lib/phone";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Search,
  UserPlus,
  Loader2,
  ChevronLeft,
  ChevronRight,
  Eye,
  Pencil,
  Trash2,
} from "lucide-react";

interface Member {
  id: number;
  name: string;
  phone: string;
  congregation: string;
  status: string;
  member_since: string | null;
  baptized: boolean;
  baptism_date: string | null;
  church_role: string;
  marital_status: string;
  origin_denomination: string;
  membership_course_completed: boolean;
  membership_course_completed_at: string | null;
  contacted: boolean;
  contacted_at: string | null;
  address: string;
  address_number: string;
  address_complement: string;
  neighborhood: string;
  city: string;
  state: string;
  created_at: string;
  updated_at: string;
}

interface PaginationMeta {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

interface MembersResponse {
  data: Member[];
  meta: PaginationMeta;
}

const maritalStatusLabels: Record<string, string> = {
  single: "Solteiro(a)",
  married: "Casado(a)",
  divorced: "Divorciado(a)",
  widowed: "Viúvo(a)",
};

const statusFilterLabels: Record<string, string> = {
  all: "Todos",
  active: "Ativo",
  inactive: "Inativo",
};

function formatDate(dateStr: string | null): string {
  if (!dateStr) return "—";
  const date = new Date(dateStr);
  return date.toLocaleDateString("pt-BR");
}

export default function MembersPage() {
  const { token } = useAuth();
  const [members, setMembers] = useState<Member[]>([]);
  const [meta, setMeta] = useState<PaginationMeta | null>(null);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("active");
  const [page, setPage] = useState(1);
  const [selectedMember, setSelectedMember] = useState<Member | null>(null);

  const fetchMembers = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (search) params.set("name", search);
      if (status !== "all") params.set("status", status);
      params.set("page", String(page));
      params.set("limit", "10");

      const res = await api<MembersResponse>(
        `/members?${params.toString()}`,
        { token }
      );
      setMembers(res.data.data);
      setMeta(res.data.meta);
    } catch {
      setMembers([]);
    } finally {
      setLoading(false);
    }
  }, [token, search, status, page]);

  async function handleDelete(id: number) {
    const confirmed = window.confirm(
      "Deseja realmente excluir este membro?"
    );

    if (!confirmed) {
      return;
    }

    try {
      await api(`/members/${id}`, {
        method: "DELETE",
        token,
      });

      fetchMembers();
    } catch (err) {
      console.error(err);
      alert("Erro ao excluir membro");
    }
  }

  useEffect(() => {
    fetchMembers();
  }, [fetchMembers]);

  function handleSearch() {
    setPage(1);
    fetchMembers();
  }

  return (
    <div className="min-w-0 space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <h1 className="text-2xl font-bold">Membros</h1>

        <div className="flex flex-wrap gap-2">
          <Button
            nativeButton={false}
            variant="outline"
            render={<Link href="/members/import" />}
          >
            Importar CSV
          </Button>

          <Button nativeButton={false} render={<Link href="/members/new" />}>
            <UserPlus className="mr-2 h-4 w-4" />
            Novo membro
          </Button>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-col gap-3 sm:flex-row">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Buscar por nome..."
            className="pl-9"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleSearch()}
          />
        </div>
        <Select
          value={status}
          onValueChange={(v) => {
            setStatus(v ?? "all");
            setPage(1);
          }}
        >
          <SelectTrigger className="w-full sm:w-40">
            <SelectValue>{statusFilterLabels[status]}</SelectValue>
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Todos</SelectItem>
            <SelectItem value="active">Ativo</SelectItem>
            <SelectItem value="inactive">Inativo</SelectItem>
          </SelectContent>
        </Select>
        <Button variant="secondary" onClick={handleSearch}>
          <Search className="mr-2 h-4 w-4" />
          Buscar
        </Button>
      </div>

      {/* Table */}
      <div className="overflow-x-auto rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Nome</TableHead>
              <TableHead>Telefone</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Congregação</TableHead>
              <TableHead>Cargo</TableHead>
              <TableHead>Membro desde</TableHead>
              <TableHead className="text-right">
                Ações
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={7} className="h-24 text-center">
                  <Loader2 className="mx-auto h-5 w-5 animate-spin text-muted-foreground" />
                </TableCell>
              </TableRow>
            ) : members.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={7}
                  className="h-24 text-center text-muted-foreground"
                >
                  Nenhum membro encontrado
                </TableCell>
              </TableRow>
            ) : (
              members.map((m) => (
                <TableRow key={m.id}>
                  <TableCell className="font-medium whitespace-nowrap">
                    <button
                      type="button"
                      className="hover:underline"
                      onClick={() => setSelectedMember(m)}
                    >
                      {m.name}
                    </button>
                  </TableCell>
                  <TableCell className="whitespace-nowrap">
                    {formatBrazilianPhone(m.phone) || "—"}
                  </TableCell>
                  <TableCell>
                    <Badge
                      variant={m.status === "active" ? "default" : "secondary"}
                    >
                      {m.status === "active" ? "Ativo" : "Inativo"}
                    </Badge>
                  </TableCell>
                  <TableCell className="whitespace-nowrap">
                    {m.congregation || "—"}
                  </TableCell>
                  <TableCell className="whitespace-nowrap">
                    {m.church_role || "—"}
                  </TableCell>
                  <TableCell className="whitespace-nowrap">
                    {formatDate(m.member_since)}
                  </TableCell>
                  <TableCell>
                    <div className="flex justify-end gap-2">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => setSelectedMember(m)}
                        aria-label={`Visualizar ${m.name}`}
                        title="Visualizar membro"
                      >
                        <Eye className="h-4 w-4" />
                      </Button>

                      <Button
                        nativeButton={false}
                        size="sm"
                        variant="outline"
                        render={<Link href={`/members/${m.id}`} />}
                      >
                        <Pencil className="h-4 w-4" />
                      </Button>

                      <Button
                        size="sm"
                        variant="destructive"
                        onClick={() => handleDelete(m.id)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      {meta && meta.total_pages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            {meta.total} membro{meta.total !== 1 && "s"} encontrado
            {meta.total !== 1 && "s"}
          </p>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="icon"
              disabled={page <= 1}
              onClick={() => setPage((p) => p - 1)}
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <span className="text-sm">
              {page} de {meta.total_pages}
            </span>
            <Button
              variant="outline"
              size="icon"
              disabled={page >= meta.total_pages}
              onClick={() => setPage((p) => p + 1)}
            >
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}

      <MemberDetailsDialog
        member={selectedMember}
        onClose={() => setSelectedMember(null)}
      />
    </div>
  );
}

function MemberDetailsDialog({
  member,
  onClose,
}: {
  member: Member | null;
  onClose: () => void;
}) {
  if (!member) return null;

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>{member.name}</DialogTitle>
          <DialogDescription>
            Informações completas do membro.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-5">
          <MemberDetailSection
            title="Dados pessoais"
            fields={[
              ["Nome", member.name],
              ["Telefone", formatBrazilianPhone(member.phone)],
              ["Status", member.status === "active" ? "Ativo" : "Inativo"],
              ["Membro desde", formatDate(member.member_since)],
              ["Estado civil", maritalStatusLabels[member.marital_status]],
            ]}
          />

          <MemberDetailSection
            title="Dados da igreja"
            fields={[
              ["Batizado", formatBoolean(member.baptized)],
              ["Data do batismo", formatDate(member.baptism_date)],
              ["Cargo na igreja", member.church_role],
              ["Congregação", member.congregation],
              ["Igreja de origem", member.origin_denomination],
              ["Curso de membresia", formatBoolean(member.membership_course_completed)],
              ["Data do curso", formatDate(member.membership_course_completed_at)],
              ["Contactado no WhatsApp", formatBoolean(member.contacted)],
              ["Data do contato", formatDate(member.contacted_at)],
            ]}
          />

          <MemberDetailSection
            title="Endereço"
            fields={[
              ["Endereço", member.address],
              ["Número", member.address_number],
              ["Complemento", member.address_complement],
              ["Bairro", member.neighborhood],
              ["Cidade", member.city],
              ["Estado", member.state],
            ]}
          />
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>Fechar</Button>
          <Button nativeButton={false} render={<Link href={`/members/${member.id}`} />}>
            <Pencil className="mr-2 h-4 w-4" />
            Editar membro
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function MemberDetailSection({
  title,
  fields,
}: {
  title: string;
  fields: Array<[string, string | undefined]>;
}) {
  return (
    <section>
      <h3 className="mb-3 font-semibold">{title}</h3>
      <dl className="grid gap-3 rounded-lg bg-muted/50 p-4 sm:grid-cols-2">
        {fields.map(([label, value]) => (
          <div key={label}>
            <dt className="text-xs text-muted-foreground">{label}</dt>
            <dd className="mt-1 text-sm font-medium">{value || "—"}</dd>
          </div>
        ))}
      </dl>
    </section>
  );
}

function formatBoolean(value: boolean): string {
  return value ? "Sim" : "Não";
}
