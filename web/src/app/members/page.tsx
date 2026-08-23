"use client";

import { useEffect, useState, useCallback } from "react";
import Link from "next/link";
import { useAuth } from "@/lib/auth";
import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
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
  CheckCircle2,
  XCircle,
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

function formatDate(dateStr: string | null): string {
  if (!dateStr) return "—";
  const date = new Date(dateStr);
  return date.toLocaleDateString("pt-BR");
}

function BoolIcon({ value }: { value: boolean }) {
  return value ? (
    <CheckCircle2 className="h-4 w-4 text-green-600" />
  ) : (
    <XCircle className="h-4 w-4 text-muted-foreground/40" />
  );
}

export default function MembersPage() {
  const { token } = useAuth();
  const [members, setMembers] = useState<Member[]>([]);
  const [meta, setMeta] = useState<PaginationMeta | null>(null);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("all");
  const [page, setPage] = useState(1);

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
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Membros</h1>

        <div className="flex gap-2">
          <Button
            variant="outline"
            render={<Link href="/members/import" />}
          >
            Importar CSV
          </Button>

          <Button render={<Link href="/members/new" />}>
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
            <SelectValue placeholder="Status" />
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
              <TableHead>Membro desde</TableHead>
              <TableHead className="text-center">Batizado</TableHead>
              <TableHead>Data do batismo</TableHead>
              <TableHead>Cargo na igreja</TableHead>
              <TableHead>Estado civil</TableHead>
              <TableHead>Congregação</TableHead>
              <TableHead>Igreja de origem</TableHead>
              <TableHead className="text-center">Curso de membresia</TableHead>
              <TableHead>Data do curso</TableHead>
              <TableHead className="text-center">Contactado no WhatsApp</TableHead>
              <TableHead>Data do contato</TableHead>
              <TableHead>Endereço</TableHead>
              <TableHead>Número</TableHead>
              <TableHead>Complemento</TableHead>
              <TableHead>Bairro</TableHead>
              <TableHead>Cidade</TableHead>
              <TableHead>Estado</TableHead>
              <TableHead className="text-right">
                Ações
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={21} className="h-24 text-center">
                  <Loader2 className="mx-auto h-5 w-5 animate-spin text-muted-foreground" />
                </TableCell>
              </TableRow>
            ) : members.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={21}
                  className="h-24 text-center text-muted-foreground"
                >
                  Nenhum membro encontrado
                </TableCell>
              </TableRow>
            ) : (
              members.map((m) => (
                <TableRow key={m.id}>
                  <TableCell className="font-medium whitespace-nowrap">
                    <Link
                      href={`/members/${m.id}`}
                      className="hover:underline"
                    >
                      {m.name}
                    </Link>
                  </TableCell>
                  <TableCell className="whitespace-nowrap">
                    {m.phone || "—"}
                  </TableCell>
                  <TableCell>
                    <Badge
                      variant={m.status === "active" ? "default" : "secondary"}
                    >
                      {m.status === "active" ? "Ativo" : "Inativo"}
                    </Badge>
                  </TableCell>
                  <TableCell className="whitespace-nowrap">
                    {formatDate(m.member_since)}
                  </TableCell>
                  <TableCell className="text-center">
                    <BoolIcon value={m.baptized} />
                  </TableCell>
                  <TableCell className="whitespace-nowrap">
                    {formatDate(m.baptism_date)}
                  </TableCell>
                  <TableCell className="whitespace-nowrap">
                    {m.church_role || "—"}
                  </TableCell>
                  <TableCell className="whitespace-nowrap">
                    {maritalStatusLabels[m.marital_status] || "—"}
                  </TableCell>
                  <TableCell className="whitespace-nowrap">
                    {m.congregation || "—"}
                  </TableCell>
                  <TableCell className="whitespace-nowrap">
                    {m.origin_denomination || "—"}
                  </TableCell>
                  <TableCell className="text-center">
                    <BoolIcon value={m.membership_course_completed} />
                  </TableCell>
                  <TableCell className="whitespace-nowrap">
                    {formatDate(m.membership_course_completed_at)}
                  </TableCell>
                  <TableCell className="text-center">
                    <BoolIcon value={m.contacted} />
                  </TableCell>
                  <TableCell className="whitespace-nowrap">
                    {formatDate(m.contacted_at)}
                  </TableCell>
                  <TableCell className="whitespace-nowrap">
                    {m.address || "—"}
                  </TableCell>
                  <TableCell className="whitespace-nowrap">
                    {m.address_number || "—"}
                  </TableCell>
                  <TableCell className="whitespace-nowrap">
                    {m.address_complement || "—"}
                  </TableCell>
                  <TableCell className="whitespace-nowrap">
                    {m.neighborhood || "—"}
                  </TableCell>
                  <TableCell className="whitespace-nowrap">
                    {m.city || "—"}
                  </TableCell>
                  <TableCell className="whitespace-nowrap">
                    {m.state || "—"}
                  </TableCell>
                  <TableCell>
                    <div className="flex justify-end gap-2">
                      <Button
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
    </div>
  );
}
